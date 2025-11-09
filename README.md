# STL Manager - Backend API

[![Tests & Quality Checks](https://github.com/vurokrazia/stl-manager-backend/actions/workflows/test.yml/badge.svg)](https://github.com/vurokrazia/stl-manager-backend/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vurokrazia/stl-manager-backend)](https://goreportcard.com/report/github.com/vurokrazia/stl-manager-backend)
[![codecov](https://codecov.io/gh/vurokrazia/stl-manager-backend/branch/main/graph/badge.svg)](https://codecov.io/gh/vurokrazia/stl-manager-backend)

Backend en Go para gestionar y clasificar archivos STL de impresión 3D usando PostgreSQL (Supabase) y OpenAI.

## Features

- 🔍 **Escaneo automático** de archivos `.stl`, `.zip`, `.rar`
- 🤖 **Clasificación IA** con OpenAI basada en nombres de archivo
- 🔎 **Búsqueda fuzzy** con PostgreSQL trigram
- 📊 **API REST** completa para gestión de archivos
- 🗄️ **PostgreSQL** (Supabase) para persistencia
- ⚡ **Chi router** - Fast, lightweight HTTP router

## Requisitos

- Go 1.22+
- PostgreSQL (Supabase)
- OpenAI API Key (opcional, para clasificación automática)

## Setup

### 1. Configurar base de datos

1. Ve a tu proyecto de Supabase: https://wkeumbfisfaawqmedofv.supabase.co
2. Abre el **SQL Editor**
3. Ejecuta el contenido de `internal/db/migrations/001_init.sql`

### 2. Configurar variables de entorno

Copia `.env.example` a `.env` y completa:

```bash
# Copia el archivo ejemplo
cp .env.example .env
```

Edita `.env` y configura:
- `DATABASE_URL` - Agrega tu password de Supabase
- `OPENAI_API_KEY` - Tu API key de OpenAI (opcional)
- `SCAN_ROOT_DIR` - Ruta de tu carpeta de STLs

### 3. Instalar dependencias

```bash
go mod download
```

### 4. Ejecutar la API

```bash
# Modo desarrollo
make dev

# O directamente
go run cmd/api/main.go
```

La API estará disponible en `http://localhost:8080`

## Uso de la API

### Autenticación

Todas las requests requieren header `X-API-Key`:

```bash
curl -H "X-API-Key: dev-secret-key" http://localhost:8080/v1/health
```

### Endpoints

#### Health Check
```bash
GET /v1/health
```

#### Escanear archivos
```bash
POST /v1/scan
X-API-Key: dev-secret-key

Response:
{
  "scan_id": "uuid"
}
```

#### Ver estado de scan
```bash
GET /v1/scans/{id}
X-API-Key: dev-secret-key

Response:
{
  "id": "uuid",
  "status": "running|completed|failed",
  "found": 100,
  "processed": 50,
  "progress": 50
}
```

#### Listar archivos
```bash
GET /v1/files?q=miata&type=stl&category=vehicle&page=1&page_size=20
X-API-Key: dev-secret-key

Response:
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

#### Obtener archivo
```bash
GET /v1/files/{id}
X-API-Key: dev-secret-key

Response:
{
  "id": "uuid",
  "path": "E:\\Impresion3D\\...",
  "file_name": "miata.stl",
  "type": "stl",
  "size": 1024000,
  "categories": ["vehicle", "rc_part"]
}
```

#### Reclasificar archivo
```bash
POST /v1/files/{id}/reclassify
X-API-Key: dev-secret-key

Response:
{
  "job_id": "uuid"
}
```

#### Listar categorías
```bash
GET /v1/categories
X-API-Key: dev-secret-key

Response:
{
  "items": [
    {"name": "figurine"},
    {"name": "vehicle"},
    ...
  ]
}
```

## Categorías disponibles

- `figurine` - Figuras/estatuas
- `miniature` - Miniaturas
- `mechanical_part` - Piezas mecánicas
- `vehicle` - Vehículos
- `anime` - Figuras de anime
- `character` - Personajes
- `diorama` - Dioramas
- `printer_upgrade` - Mejoras para impresora
- `tool_holder` - Portaherramientas
- `rc_part` - Piezas RC
- `uncategorized` - Sin categoría

## Desarrollo

```bash
# Formatear código
make fmt

# Ejecutar tests
make test

# Compilar binario
make build

# Ver todos los comandos
make help
```

### Tests

El proyecto incluye tests de integración completos para todos los endpoints.

```bash
# Ejecutar todos los tests
go test ./tests/integration/... -v

# Ejecutar tests de un módulo específico
go test ./tests/integration/categories/... -v

# Ejecutar con cobertura
go test ./tests/integration/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Tests en paralelo
go test ./tests/integration/... -v -parallel 4
```

**Cobertura actual:**
- ✅ Categories API - 17 tests
- ✅ Health API - 2 tests
- ✅ Browse API - 3 tests
- ✅ Folders API - 4 tests
- ✅ Files API - 2 tests
- ✅ Scans API - 2 tests

Ver más en [tests/README.md](tests/README.md)

## Estructura del proyecto

```
stl-manager-backend/
├── cmd/
│   └── api/          # Entrypoint de la API
├── internal/
│   ├── ai/           # Cliente OpenAI para clasificación
│   ├── config/       # Configuración
│   ├── db/           # Base de datos y migraciones
│   ├── handlers/     # HTTP handlers
│   └── scanner/      # Scanner de archivos
├── .env              # Variables de entorno
├── Makefile          # Comandos útiles
└── README.md         # Este archivo
```

## Troubleshooting

### Error: "database unhealthy"
- Verifica que `DATABASE_URL` en `.env` tenga el password correcto
- Verifica que puedas conectarte a Supabase

### Error: "failed to walk directory"
- Verifica que `SCAN_ROOT_DIR` exista y sea accesible
- En Windows, usa doble backslash: `E:\\Impresion3D`

### No clasifica archivos
- Agrega tu `OPENAI_API_KEY` en `.env`
- Sin API key, los archivos quedan sin categoría

## TODO

- [ ] Implementar workers con goroutines para clasificación async
- [ ] Agregar paginación completa en todos los endpoints
- [ ] Implementar búsqueda trigram completa
- [ ] Agregar tests unitarios
- [ ] Agregar métricas y observabilidad
- [ ] CLI tool para uso local sin API

## Licencia

MIT
