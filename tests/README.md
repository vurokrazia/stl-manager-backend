# Tests - STL Manager Backend API

Suite completa de tests de integración para todos los endpoints del API.

## 📁 Estructura

```
tests/
├── integration/              # Tests de integración con DB real
│   ├── helpers/             # Helpers reutilizables DRY
│   │   ├── database.go      # DB setup, CRUD helpers
│   │   ├── http.go          # HTTP request/response helpers
│   │   └── assertions.go    # Assertions comunes
│   └── categories/          # Tests de categories API
│       ├── setup_test.go    # Setup compartido
│       ├── list_test.go     # GET /v1/categories
│       ├── create_test.go   # POST /v1/categories
│       ├── get_test.go      # GET /v1/categories/{id}
│       ├── update_test.go   # PUT /v1/categories/{id}
│       ├── delete_test.go   # DELETE /v1/categories/{id}
│       └── restore_test.go  # POST /v1/categories/{id}/restore
└── README.md                # Este archivo
```

## 🚀 Ejecutar Tests

### Todos los tests
```bash
go test ./tests/integration/... -v
```

### Solo categories
```bash
go test ./tests/integration/categories/... -v
```

### Test específico
```bash
go test ./tests/integration/categories/... -v -run TestListCategories
```

### Con cobertura
```bash
# Cobertura de un módulo
go test ./tests/integration/categories/... -cover

# Reporte detallado
go test ./tests/integration/categories/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Sin caché (forzar re-ejecución)
```bash
go test ./tests/integration/... -v -count=1
```

### Tests en paralelo
```bash
go test ./tests/integration/... -v -parallel 4
```

## ✅ Casos de Uso Cubiertos

### Categories API

#### GET /v1/categories (3 tests)
- ✅ Listar todas las categorías (paginación default)
- ✅ Listar con paginación custom (`?page=1&page_size=5`)
- ✅ Buscar categorías (`?q=search`)

#### POST /v1/categories (3 tests)
- ✅ Crear categoría exitosamente
- ✅ Validación: nombre vacío falla
- ✅ Validación: JSON inválido falla

#### GET /v1/categories/{id} (3 tests)
- ✅ Obtener categoría existente por ID
- ✅ Error: ID inválido (400 Bad Request)
- ✅ Error: categoría no encontrada (404 Not Found)

#### PUT /v1/categories/{id} (3 tests)
- ✅ Actualizar nombre exitosamente
- ✅ Validación: nombre vacío falla
- ✅ Error: ID inválido

#### DELETE /v1/categories/{id} (3 tests)
- ✅ Soft delete exitoso
- ✅ Error: ID inválido
- ✅ Verificar: categoría eliminada NO aparece en listados

#### POST /v1/categories/{id}/restore (2 tests)
- ✅ Restaurar categoría eliminada
- ✅ Error: ID inválido

**Total Categories: 17 tests**

---

## 🛠️ Helpers Reutilizables (DRY)

### Database Helpers (`helpers/database.go`)

```go
// Setup/Cleanup
SetupTestDatabase()           // Conecta a DB
CleanupTestDatabase()         // Cierra conexión

// Category CRUD
CreateTestCategory(t, name)   // Crea categoría de test
DeleteTestCategory(t, id)     // Limpia categoría (hard delete)
SoftDeleteTestCategory(t, id) // Soft delete
RestoreTestCategory(t, id)    // Restaura categoría
GetTestCategory(t, id)        // Obtiene categoría
```

### HTTP Helpers (`helpers/http.go`)

```go
// Request builders
GET(url)                      // GET request
POST(url, body)               // POST request
PUT(url, body)                // PUT request
PATCH(url, body)              // PATCH request
DELETE(url)                   // DELETE request

// Fluent API
req.WithURLParam("id", "123")
req.WithQueryParam("q", "search")
req.WithHeader("X-API-Key", "key")

// Response helpers
resp.GetString("field")
resp.GetFloat("field")
resp.GetArray("items")
resp.GetMap("data")

// Ejecutar request
MakeRequest(t, req, handler)
```

### Assertion Helpers (`helpers/assertions.go`)

```go
AssertSuccessResponse(t, resp, 200)
AssertErrorResponse(t, resp, 400)
AssertHasFields(t, body, "id", "name")
AssertPaginatedResponse(t, resp)
```

---

## 📝 Ejemplo de Uso

```go
package categories

import (
    "testing"
    "net/http"
    "stl-manager/tests/integration/helpers"
)

func TestListCategories(t *testing.T) {
    // Crear datos de prueba
    cat := helpers.CreateTestCategory(t, "test-category")
    defer helpers.DeleteTestCategory(t, cat.ID)

    // Hacer request
    req := helpers.GET("/categories").
        WithQueryParam("q", "test").
        WithQueryParam("page", "1")

    resp := helpers.MakeRequest(t, req, handler.ListCategories)

    // Assertions
    helpers.AssertSuccessResponse(t, resp, http.StatusOK)
    helpers.AssertPaginatedResponse(t, resp)

    items := resp.GetArray("items")
    assert.GreaterOrEqual(t, len(items), 1)
}
```

---

## 🔧 Configuración

### Requisitos

1. **Go 1.21+**
2. **PostgreSQL** (Supabase configurado en `.env`)
3. **Dependencias**:
   ```bash
   go get github.com/stretchr/testify
   go get github.com/go-chi/chi/v5
   ```

### Variables de Entorno

Archivo `.env` en la raíz del proyecto:

```env
DATABASE_URL=postgresql://user:password@host:5432/database
```

---

## 📊 Resultados

### Output de Ejecución

```
=== RUN   TestListCategories
=== RUN   TestListCategories/list_all_categories
=== RUN   TestListCategories/list_with_pagination
=== RUN   TestListCategories/search_categories
--- PASS: TestListCategories (1.02s)
=== RUN   TestCreateCategory
--- PASS: TestCreateCategory (1.03s)
=== RUN   TestGetCategory
--- PASS: TestGetCategory (0.43s)
=== RUN   TestUpdateCategory
--- PASS: TestUpdateCategory (0.39s)
=== RUN   TestSoftDeleteCategory
--- PASS: TestSoftDeleteCategory (0.34s)
=== RUN   TestRestoreCategory
--- PASS: TestRestoreCategory (0.43s)
=== RUN   TestSoftDeleteHidesCategory
--- PASS: TestSoftDeleteHidesCategory (0.77s)
PASS
ok  	stl-manager/tests/integration/categories	5.224s
```

---

## 🎯 Mejores Prácticas

### 1. Nombres Únicos
Todos los tests crean datos con UUIDs únicos para evitar colisiones:
```go
cat := helpers.CreateTestCategory(t, "test-list") // Agrega UUID automático
```

### 2. Cleanup con defer
Siempre limpia datos de prueba:
```go
cat := helpers.CreateTestCategory(t, "test")
defer helpers.DeleteTestCategory(t, cat.ID)
```

### 3. Table-Driven Tests
Usa subtests para organizar casos:
```go
tests := []struct {
    name     string
    req      helpers.HTTPTestRequest
    wantCode int
}{
    {name: "success", req: helpers.GET("/"), wantCode: 200},
    {name: "not found", req: helpers.GET("/404"), wantCode: 404},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        resp := helpers.MakeRequest(t, tt.req, handler)
        assert.Equal(t, tt.wantCode, resp.Code)
    })
}
```

### 4. Tests Independientes
Cada test debe poder ejecutarse solo:
```bash
go test ./tests/integration/categories/... -v -run TestGetCategory
```

---

## 🐛 Troubleshooting

### Error: DATABASE_URL not set
Asegúrate de tener `.env` en la raíz:
```bash
DATABASE_URL=postgresql://...
```

### Tests fallan con "connection refused"
Verifica que PostgreSQL esté corriendo.

### Tests crean datos duplicados
Los UUIDs únicos deberían prevenir esto. Si ocurre:
```bash
go test ./tests/integration/... -count=1
```

---

## 📈 Tests Completados

- [x] **Categories API** - 17 tests ✅
- [x] **Files API** - 15 tests ✅
- [x] **Folders API** - 17 tests ✅
- [x] **Browse API** - 5 tests ✅
- [x] **Scans API** - 10 tests ✅
- [x] **Health API** - 2 tests ✅

**TOTAL: 66 tests - TODOS PASANDO** ✅

### Tiempo de Ejecución
```
browse:      6.564s
categories:  5.292s
files:       7.676s
folders:    13.745s
health:      2.515s
scans:       3.336s
─────────────────────
TOTAL:     ~39 seconds
```

---

## 🤝 Contribuir

### Agregar nuevos tests

1. **Crear archivo**: `tests/integration/<module>/<endpoint>_test.go`
2. **Usar helpers**: Reutiliza `helpers/` para DRY
3. **Seguir patrón**: Table-driven tests con subtests
4. **Cleanup**: Siempre usar `defer` para limpiar
5. **Ejecutar**: `go test ./tests/integration/... -v`

### Ejemplo

```go
// tests/integration/files/list_test.go
package files

import (
    "testing"
    "net/http"
    "stl-manager/tests/integration/helpers"
)

func TestListFiles(t *testing.T) {
    req := helpers.GET("/files").WithQueryParam("page", "1")
    resp := helpers.MakeRequest(t, req, handler.ListFiles)

    helpers.AssertSuccessResponse(t, resp, http.StatusOK)
    helpers.AssertPaginatedResponse(t, resp)
}
```

---

**Última actualización**: 2024-11-08
