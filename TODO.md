# TODO - Optimizaciones de Velocidad API

**Última actualización:** 2025-11-09
**Estado:** Análisis completado - Pendiente implementación

---

## 📋 Índice

- [Optimizaciones YA implementadas](#-optimizaciones-ya-implementadas)
- [Problemas CRÍTICOS](#-problemas-críticos)
- [Mejoras MEDIAS](#-mejoras-medias)
- [Mejoras MENORES](#-mejoras-menores)
- [Priorización](#-priorización-por-impacto)
- [Plan de implementación](#-plan-de-implementación)

---

## ✅ Optimizaciones YA implementadas

Estas optimizaciones ya están funcionando correctamente:

- [x] **Batch queries para categorías** - `GetCategoriesBatch`, `GetFolderCategoriesBatch` (1 query en lugar de N)
- [x] **Bulk operations** - `BulkAddFileCategories`, `BulkRemoveFolderCategories`
- [x] **Paginación** - Todos los endpoints list tienen paginación
- [x] **Paralelización en scans** - 20 workers concurrentes (`internal/handlers/scans/run.go:86`)
- [x] **CountFolderFiles** - Uso de COUNT en lugar de cargar todos los registros

---

## 🔴 Problemas CRÍTICOS

### 1. N+1 queries en bucles (ALTO IMPACTO)

**Impacto:** 50-90% reducción en tiempo de respuesta
**Esfuerzo:** 2-3 horas

**Archivos afectados:**
- `internal/handlers/browse/all.go:136` - CountFolderFiles dentro del loop
- `internal/handlers/browse/all.go:299` - Mismo patrón
- `internal/handlers/folders/all.go:133` - Mismo patrón
- `internal/handlers/folders/all.go:340` - Mismo patrón

**Problema:**
```go
for _, folder := range folders {
    fileCount, err := queries.CountFolderFiles(ctx, folder.ID)  // ❌ N queries
    // ...
}
```

**Solución:**
Crear query batch para counts en `internal/db/queries/folders.sql`:

```sql
-- name: GetFolderFileCountsBatch :many
SELECT folder_id, COUNT(*) as file_count
FROM files
WHERE folder_id = ANY($1::uuid[])
GROUP BY folder_id;
```

Luego modificar los handlers para usar batch:
```go
// Collect folder IDs
folderIDs := make([]pgtype.UUID, len(folders))
for i, folder := range folders {
    folderIDs[i] = folder.ID
}

// Get all counts in ONE query
countsMap := make(map[pgtype.UUID]int64)
if len(folderIDs) > 0 {
    batchCounts, err := queries.GetFolderFileCountsBatch(ctx, folderIDs)
    if err != nil {
        h.logger.Warn("failed to get file counts batch", zap.Error(err))
    } else {
        for _, row := range batchCounts {
            countsMap[row.FolderID] = row.FileCount
        }
    }
}

// Use cached counts
for _, folder := range folders {
    fileCount := countsMap[folder.ID]
    // ...
}
```

**Archivos a modificar:**
- [ ] `internal/db/queries/folders.sql` - Agregar query batch
- [ ] `internal/handlers/browse/all.go:136` - Usar batch query
- [ ] `internal/handlers/browse/all.go:299` - Usar batch query
- [ ] `internal/handlers/folders/all.go:133` - Usar batch query
- [ ] `internal/handlers/folders/all.go:340` - Usar batch query
- [ ] Regenerar código con `sqlc generate`

---

### 2. Filtrado en memoria (ALTO IMPACTO)

**Impacto:** 80-95% mejora en uso de memoria y tiempo
**Esfuerzo:** 3-4 horas

**Archivo afectado:** `internal/handlers/folders/all.go:238-283`

**Problema:**
```go
if searchQuery != "" || categoryFilter != "" {
    allFiles, err := queries.GetFolderFiles(ctx, folder.ID)  // ❌ Carga TODOS

    filtered := []db.File{}
    for _, file := range allFiles {
        if searchQuery != "" && !strings.Contains(...) {  // ❌ Filtrado en memoria
            continue
        }
        if typeFilter != "" && strings.ToLower(file.Type) != typeFilter {
            continue
        }
        // ...
    }
}
```

**Solución:**
Crear queries SQL específicas en `internal/db/queries/files.sql`:

```sql
-- name: GetFolderFilesBySearchAndType :many
SELECT f.* FROM files f
WHERE f.folder_id = $1
  AND ($2 = '' OR f.file_name ILIKE '%' || $2 || '%')
  AND ($3 = '' OR f.type = $4)
ORDER BY f.file_name
LIMIT $5 OFFSET $6;

-- name: CountFolderFilesBySearchAndType :one
SELECT COUNT(*) FROM files f
WHERE f.folder_id = $1
  AND ($2 = '' OR f.file_name ILIKE '%' || $2 || '%')
  AND ($3 = '' OR f.type = $4);

-- name: GetFolderFilesByCategory :many
SELECT f.* FROM files f
INNER JOIN files_categories fc ON f.id = fc.file_id
INNER JOIN categories c ON fc.category_id = c.id
WHERE f.folder_id = $1
  AND c.name = $2
ORDER BY f.file_name
LIMIT $3 OFFSET $4;

-- name: GetFolderFilesBySearchTypeAndCategory :many
SELECT DISTINCT f.* FROM files f
INNER JOIN files_categories fc ON f.id = fc.file_id
INNER JOIN categories c ON fc.category_id = c.id
WHERE f.folder_id = $1
  AND ($2 = '' OR f.file_name ILIKE '%' || $2 || '%')
  AND ($3 = '' OR f.type = $4)
  AND c.name = $5
ORDER BY f.file_name
LIMIT $6 OFFSET $7;
```

**Mismo patrón para filtrado de subfolders:**
```sql
-- name: GetSubfoldersByCategory :many
SELECT DISTINCT fo.* FROM folders fo
INNER JOIN folders_categories fc ON fo.id = fc.folder_id
INNER JOIN categories c ON fc.category_id = c.id
WHERE fo.parent_folder_id = $1
  AND c.name = $2
ORDER BY fo.name;
```

**Archivos a modificar:**
- [ ] `internal/db/queries/files.sql` - Agregar queries con filtros
- [ ] `internal/db/queries/folders.sql` - Agregar queries de subfolders con filtros
- [ ] `internal/handlers/folders/all.go:205-231` - Reemplazar filtrado en memoria de subfolders
- [ ] `internal/handlers/folders/all.go:238-302` - Reemplazar filtrado en memoria de files
- [ ] Regenerar código con `sqlc generate`

---

### 3. Búsqueda con ILIKE sin índice (MEDIO IMPACTO)

**Impacto:** 10-50x más rápidas en búsquedas de texto
**Esfuerzo:** 1 hora

**Archivos afectados:**
- `internal/db/queries/folders.sql` - `SearchFoldersPaginated`, `SearchRootFoldersPaginated`
- `internal/db/queries/files.sql` - `SearchFiles`

**Problema:**
```sql
WHERE name ILIKE '%' || $3::text || '%'  -- ❌ No puede usar índice B-tree
```

**Solución:**
Crear nueva migración con extensión pg_trgm e índices GIN:

```sql
-- Migration: migrations/009_add_text_search_indexes.sql

-- Up Migration
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Índices trigram para búsqueda de texto completo
CREATE INDEX idx_folders_name_trgm ON folders USING gin(name gin_trgm_ops);
CREATE INDEX idx_files_name_trgm ON files USING gin(file_name gin_trgm_ops);
CREATE INDEX idx_files_path_trgm ON files USING gin(path gin_trgm_ops);

-- Índice para búsqueda por similaridad
CREATE INDEX idx_folders_name_similarity ON folders USING gist(name gist_trgm_ops);
CREATE INDEX idx_files_name_similarity ON files USING gist(file_name gist_trgm_ops);

-- Down Migration
-- DROP INDEX IF EXISTS idx_folders_name_trgm;
-- DROP INDEX IF EXISTS idx_files_name_trgm;
-- DROP INDEX IF EXISTS idx_files_path_trgm;
-- DROP INDEX IF EXISTS idx_folders_name_similarity;
-- DROP INDEX IF EXISTS idx_files_name_similarity;
-- DROP EXTENSION IF EXISTS pg_trgm;
```

**Archivos a crear/modificar:**
- [ ] `migrations/009_add_text_search_indexes.sql` - Nueva migración
- [ ] Documentar en README que se requiere extensión `pg_trgm`

---

## 🟡 Mejoras MEDIAS

### 4. Falta de índices compuestos

**Impacto:** 20-40% más rápido en queries complejas
**Esfuerzo:** 1 hora

**Solución:**
Crear nueva migración con índices compuestos:

```sql
-- Migration: migrations/010_add_composite_indexes.sql

-- Up Migration
-- Índices compuestos para files
CREATE INDEX idx_files_folder_id_name ON files(folder_id, file_name);
CREATE INDEX idx_files_folder_id_type ON files(folder_id, type);
CREATE INDEX idx_files_type ON files(type);

-- Índices para jerarquía de folders
CREATE INDEX idx_folders_parent_folder_id ON folders(parent_folder_id);
CREATE INDEX idx_folders_parent_name ON folders(parent_folder_id, name);

-- Índices para relaciones many-to-many
CREATE INDEX idx_files_categories_file_id ON files_categories(file_id);
CREATE INDEX idx_files_categories_category_id ON files_categories(category_id);
CREATE INDEX idx_folders_categories_folder_id ON folders_categories(folder_id);
CREATE INDEX idx_folders_categories_category_id ON folders_categories(category_id);

-- Down Migration
-- DROP INDEX IF EXISTS idx_files_folder_id_name;
-- DROP INDEX IF EXISTS idx_files_folder_id_type;
-- DROP INDEX IF EXISTS idx_files_type;
-- DROP INDEX IF EXISTS idx_folders_parent_folder_id;
-- DROP INDEX IF EXISTS idx_folders_parent_name;
-- DROP INDEX IF EXISTS idx_files_categories_file_id;
-- DROP INDEX IF EXISTS idx_files_categories_category_id;
-- DROP INDEX IF EXISTS idx_folders_categories_folder_id;
-- DROP INDEX IF EXISTS idx_folders_categories_category_id;
```

**Archivos a crear:**
- [ ] `migrations/010_add_composite_indexes.sql` - Nueva migración

---

### 5. Sin caching de resultados

**Impacto:** 50-70% reducción en carga de DB
**Esfuerzo:** 6-8 horas

**Endpoints a cachear:**

| Endpoint | TTL | Invalidar en |
|----------|-----|--------------|
| `GET /v1/categories` | 5 min | POST/PUT/DELETE categories |
| `GET /v1/folders` (sin params) | 2 min | POST /v1/scan completed |
| `GET /v1/browse` (sin params) | 2 min | POST /v1/scan completed |
| Counts de archivos por folder | 1 min | POST /v1/scan completed |

**Solución:**
1. Agregar Redis como dependencia:
```bash
go get github.com/redis/go-redis/v9
```

2. Crear servicio de cache:
```go
// internal/cache/redis.go
package cache

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func New(addr string) *RedisCache {
    return &RedisCache{
        client: redis.NewClient(&redis.Options{
            Addr: addr,
        }),
    }
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, pattern string) error {
    // Delete by pattern
    iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        c.client.Del(ctx, iter.Val())
    }
    return iter.Err()
}
```

3. Middleware de cache:
```go
// internal/middleware/cache.go
package middleware

func CacheMiddleware(cache *cache.RedisCache, ttl time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Solo cachear GET sin query params
            if r.Method != "GET" || r.URL.RawQuery != "" {
                next.ServeHTTP(w, r)
                return
            }

            cacheKey := "api:" + r.URL.Path

            // Try to get from cache
            cached, err := cache.Get(r.Context(), cacheKey)
            if err == nil {
                w.Header().Set("Content-Type", "application/json")
                w.Header().Set("X-Cache", "HIT")
                w.Write([]byte(cached))
                return
            }

            // Cache miss - capture response
            rec := httptest.NewRecorder()
            next.ServeHTTP(rec, r)

            // Store in cache
            if rec.Code == 200 {
                cache.Set(r.Context(), cacheKey, rec.Body.String(), ttl)
            }

            // Write response
            for k, v := range rec.Header() {
                w.Header()[k] = v
            }
            w.Header().Set("X-Cache", "MISS")
            w.WriteHeader(rec.Code)
            w.Write(rec.Body.Bytes())
        })
    }
}
```

**Archivos a crear/modificar:**
- [ ] `internal/cache/redis.go` - Nuevo servicio de cache
- [ ] `internal/middleware/cache.go` - Middleware de cache
- [ ] `internal/config/config.go` - Agregar `RedisURL string`
- [ ] `cmd/api/main.go` - Inicializar Redis y aplicar middleware
- [ ] `.env.example` - Agregar `REDIS_URL=redis://localhost:6379`
- [ ] `go.mod` - Agregar dependencia de redis

---

### 6. Connection pool no configurado optimalmente

**Impacto:** 10-20% mejor throughput bajo concurrencia
**Esfuerzo:** 30 minutos

**Archivo afectado:** `cmd/api/main.go:49-55`

**Solución:**
```go
// Configurar pool explícitamente
poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
if err != nil {
    logger.Fatal("failed to parse database URL", zap.Error(err))
}

// Configuración optimizada para API
poolConfig.MaxConns = 50                    // Máximo de conexiones
poolConfig.MinConns = 10                    // Conexiones mínimas idle
poolConfig.MaxConnLifetime = time.Hour      // Reciclar conexiones cada hora
poolConfig.MaxConnIdleTime = 30 * time.Minute
poolConfig.HealthCheckPeriod = time.Minute

// Configuración de prepared statements
poolConfig.ConnConfig.RuntimeParams["application_name"] = "stl-manager-api"
poolConfig.ConnConfig.PreferSimpleProtocol = false  // Usar prepared statements

pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
```

**Archivos a modificar:**
- [ ] `cmd/api/main.go:49-55` - Configurar pool explícitamente

---

### 7. Sin compresión HTTP

**Impacto:** 60-80% reducción en tamaño de respuesta
**Esfuerzo:** 30 minutos

**Archivo afectado:** `cmd/api/main.go:76-95`

**Solución:**
```go
import "github.com/go-chi/chi/v5/middleware"

// Agregar middleware de compresión
r.Use(middleware.Compress(5, "application/json"))  // Nivel 5 de gzip
```

**Configuración adicional:**
```go
// Solo comprimir responses > 1KB
r.Use(middleware.Compress(5, "application/json", "text/plain"))
```

**Archivos a modificar:**
- [ ] `cmd/api/main.go:79` - Agregar `r.Use(middleware.Compress(5))`

---

## 🟢 Mejoras MENORES

### 8. Timeouts generosos

**Impacto:** Mejor gestión de recursos
**Esfuerzo:** 15 minutos

**Archivo afectado:** `cmd/api/main.go:84`

**Solución:**
Timeouts por tipo de endpoint:
```go
// Default timeout más corto
r.Use(middleware.Timeout(10 * time.Second))

// Rutas con timeouts específicos
r.Route("/v1", func(r chi.Router) {
    // Endpoints rápidos (listados sin filtros)
    r.Group(func(r chi.Router) {
        r.Use(middleware.Timeout(5 * time.Second))
        r.Get("/health", baseHandler.Health)
        r.Get("/categories", categoriesHandler.ListCategories)
    })

    // Endpoints con búsquedas
    r.Group(func(r chi.Router) {
        r.Use(middleware.Timeout(15 * time.Second))
        r.Get("/files", filesHandler.ListFiles)
        r.Get("/folders", foldersHandler.ListFolders)
        r.Get("/browse", browseHandler.ListBrowse)
    })

    // Endpoints pesados (scans ya son async)
    r.Post("/scan", scansHandler.CreateScan)
})
```

**Archivos a modificar:**
- [ ] `cmd/api/main.go:84` - Implementar timeouts por grupo de rutas

---

### 9. Queries de count duplicados

**Impacto:** Eliminar 1 query por request
**Esfuerzo:** 2-3 horas

**Solución:**
Usar CTEs (Common Table Expressions) para combinar SELECT + COUNT:

```sql
-- name: ListFilesWithCount :many
WITH counted AS (
  SELECT
    *,
    COUNT(*) OVER() as total_count
  FROM files
  WHERE ($1 = '' OR file_name ILIKE '%' || $1 || '%')
    AND ($2 = '' OR type = $2)
  ORDER BY file_name ASC
)
SELECT * FROM counted
LIMIT $3 OFFSET $4;
```

Cambiar handler para leer el count del primer row:
```go
rows, err := queries.ListFilesWithCount(ctx, params)
var total int64
if len(rows) > 0 {
    total = rows[0].TotalCount
} else {
    total = 0
}
```

**Archivos a modificar:**
- [ ] `internal/db/queries/files.sql` - Agregar queries con CTE
- [ ] `internal/db/queries/folders.sql` - Agregar queries con CTE
- [ ] Handlers correspondientes - Leer count del resultado

---

### 10. Sin prepared statements explícitos

**Impacto:** Pequeña mejora en queries repetitivos
**Esfuerzo:** 5 minutos

**Solución:**
Ya incluido en #6 (Connection pool config):
```go
poolConfig.ConnConfig.PreferSimpleProtocol = false  // Forzar prepared statements
```

---

## 📊 Priorización por IMPACTO

| Prioridad | # | Optimización | Impacto estimado | Esfuerzo |
|-----------|---|--------------|------------------|----------|
| 🔴 **P0** | 1 | Batch COUNT queries | 50-90% mejora | 2-3 horas |
| 🔴 **P0** | 2 | Filtrado en SQL | 80-95% mejora | 3-4 horas |
| 🟡 **P1** | 4 | Índices compuestos | 20-40% mejora | 1 hora |
| 🟡 **P1** | 3 | pg_trgm para búsquedas | 10-50x mejora | 1 hora |
| 🟡 **P1** | 7 | HTTP compression | 60-80% menos ancho de banda | 30 min |
| 🟡 **P1** | 6 | Connection pool config | 10-20% más throughput | 30 min |
| 🟡 **P2** | 5 | Redis caching | 50-70% menos carga DB | 6-8 horas |
| 🟢 **P3** | 8 | Timeouts por endpoint | Mejor gestión recursos | 15 min |
| 🟢 **P3** | 9 | CTEs para counts | Eliminar 1 query/request | 2-3 horas |
| 🟢 **P3** | 10 | Prepared statements | Pequeña mejora | 5 min |

---

## 🎯 Plan de implementación

### FASE 1: Quick Wins (1 día de trabajo)

**Objetivo:** Mejoras rápidas con alto impacto

- [ ] **#7** - HTTP compression (30 min)
- [ ] **#6** - Connection pool config (30 min)
- [ ] **#10** - Prepared statements (5 min)
- [ ] **#4** - Índices compuestos (1 hora)
- [ ] **#3** - pg_trgm para búsquedas (1 hora)
- [ ] **#1** - Batch COUNT queries (2-3 horas)

**Resultado esperado:** ~40-60% mejora general en velocidad

---

### FASE 2: Optimizaciones SQL (2 días de trabajo)

**Objetivo:** Eliminar queries ineficientes

- [ ] **#2** - Filtrado en SQL en lugar de memoria (3-4 horas)
- [ ] **#9** - CTEs para eliminar COUNT duplicados (2-3 horas)
- [ ] **#8** - Timeouts por endpoint (15 min)
- [ ] Testing y benchmarking

**Resultado esperado:** 70-85% mejora general desde baseline

---

### FASE 3: Arquitectura avanzada (1 semana de trabajo)

**Objetivo:** Escalabilidad y resiliencia

- [ ] **#5** - Redis caching (6-8 horas)
- [ ] Rate limiting por endpoint (4 horas)
- [ ] Metrics con Prometheus (4 horas)
- [ ] Distributed tracing con OpenTelemetry (8 horas)
- [ ] Load testing con k6 (4 horas)

**Resultado esperado:** Sistema preparado para escalar horizontalmente

---

## 📝 Notas y consideraciones

### Antes de implementar

1. **Backup de base de datos** antes de aplicar nuevas migraciones
2. **Benchmarking baseline** - Medir performance actual con herramienta como `hey` o `k6`:
   ```bash
   # Ejemplo de benchmark
   hey -n 1000 -c 10 -H "X-API-Key: your-key" http://localhost:8080/v1/folders
   ```
3. **Configurar ambiente de staging** para probar optimizaciones

### Métricas a monitorear

- **Response time (p50, p95, p99)**
- **Throughput (requests/segundo)**
- **Database query time**
- **Connection pool usage**
- **Memory usage**
- **CPU usage**

### Herramientas recomendadas

- **Benchmarking:** `k6`, `hey`, `wrk`
- **Profiling:** `pprof` (ya incluido en Go)
- **Monitoring:** Prometheus + Grafana
- **APM:** Jaeger o Tempo para tracing

---

## 🔗 Referencias

- [pgx performance guide](https://github.com/jackc/pgx/wiki/Performance-and-Memory-Usage)
- [PostgreSQL pg_trgm](https://www.postgresql.org/docs/current/pgtrgm.html)
- [Chi middleware](https://github.com/go-chi/chi#middlewares)
- [Redis caching patterns](https://redis.io/docs/manual/patterns/)

---

## ❓ Preguntas pendientes

Para priorizar correctamente, necesitamos:

1. **¿Cuál es el volumen actual de datos?**
   - Número de archivos en DB: ___?
   - Número de folders: ___?
   - Número de categorías: ___?

2. **¿Cuáles endpoints son los más lentos actualmente?**
   - `/v1/browse`?
   - `/v1/folders`?
   - `/v1/files`?

3. **¿Tienes Redis disponible** para implementar caching?
   - [ ] Sí, ya instalado
   - [ ] Sí, puedo instalar
   - [ ] No, no es opción

4. **¿Cuál es tu prioridad máxima?**
   - [ ] a) Reducir latencia de endpoints específicos
   - [ ] b) Soportar más concurrencia
   - [ ] c) Reducir carga de base de datos
   - [ ] d) Todas las anteriores

---

**Última revisión:** 2025-11-09
