# STL Manager API - Documentation

**Base URL**: `http://localhost:8081/v1` (development)

**Autenticación**: Todos los endpoints requieren el header `X-API-Key`

**Última actualización**: 2024-11-02

---

## Índice de Endpoints

### Health & Status
- [GET /v1/health](#get-v1health) - Health check
- [GET /v1/ai/status](#get-v1aistatus) - Estado de clasificación AI

### Scans
- [POST /v1/scan](#post-v1scan) - Crear nuevo scan
- [GET /v1/scans](#get-v1scans) - Listar scans
- [GET /v1/scans/{id}](#get-v1scansid) - Obtener scan por ID

### Files
- [GET /v1/files](#get-v1files) - Listar archivos
- [GET /v1/files/{id}](#get-v1filesid) - Obtener archivo por ID
- [POST /v1/files/{id}/reclassify](#post-v1filesidReclassify) - Reclasificar archivo
- [PATCH /v1/files/{id}/categories](#patch-v1filesidcategories) - Actualizar categorías de archivo

### Categories
- [GET /v1/categories](#get-v1categories) - Listar categorías
- [POST /v1/categories](#post-v1categories) - Crear categoría
- [GET /v1/categories/{id}](#get-v1categoriesid) - Obtener categoría por ID
- [PUT /v1/categories/{id}](#put-v1categoriesid) - Actualizar categoría
- [DELETE /v1/categories/{id}](#delete-v1categoriesid) - Eliminar categoría (soft delete)
- [POST /v1/categories/{id}/restore](#post-v1categoriesidrestore) - Restaurar categoría eliminada

### Browse & Navigation
- [GET /v1/browse](#get-v1browse) - Navegar folders raíz
- [GET /v1/mixed](#get-v1mixed) - Vista mixta (folders + archivos)

### Folders
- [GET /v1/folders](#get-v1folders) - Listar folders
- [GET /v1/folders/{id}](#get-v1foldersid) - Obtener folder con contenido
- [PATCH /v1/folders/{id}/categories](#patch-v1foldersidcategories) - Actualizar categorías de folder

---

## Health & Status

### GET /v1/health

**Descripción**: Verifica el estado de salud de la API y la conexión a la base de datos

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/health`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```

**Response Success (200 OK):**
```json
{
  "status": "healthy",
  "service": "stl-manager-api"
}
```

**Response Error (503 Service Unavailable):**
```json
{
  "error": "database unhealthy"
}
```

**Códigos de estado:**
- `200`: Servicio operativo
- `503`: Base de datos no disponible

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8081/v1/health \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/ai/status

**Descripción**: Verifica si la clasificación automática con OpenAI está habilitada

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/ai/status`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```

**Response Success (200 OK):**
```json
{
  "enabled": true
}
```

**Códigos de estado:**
- `200`: Status obtenido exitosamente

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8081/v1/ai/status \
  -H "X-API-Key: dev-secret-key"
```

---

## Scans

### POST /v1/scan

**Descripción**: Crea un nuevo scan del sistema de archivos. El proceso se ejecuta en segundo plano y escanea el directorio configurado (`SCAN_ROOT_DIR`) buscando archivos STL, ZIP y RAR.

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: POST
- **URL**: `/v1/scan`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```

**Response Success (202 Accepted):**
```json
{
  "scan_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": "failed to create scan"
}
```

**Códigos de estado:**
- `202`: Scan iniciado correctamente
- `500`: Error al crear el scan

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8081/v1/scan \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/scans

**Descripción**: Lista todos los scans con paginación, ordenados por fecha de creación (más recientes primero)

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/scans`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "completed",
      "found": 150,
      "processed": 150,
      "progress": 100,
      "error": "",
      "created_at": "2024-11-02T10:30:00Z",
      "updated_at": "2024-11-02T10:35:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 20
}
```

**Valores de status:**
- `running`: Scan en progreso
- `completed`: Scan completado exitosamente
- `failed`: Scan falló (ver campo `error`)

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `500`: Error al listar scans

**Ejemplo con cURL:**
```bash
curl -X GET "http://localhost:8081/v1/scans?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/scans/{id}

**Descripción**: Obtiene el detalle de un scan específico por su ID

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/scans/{id}`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del scan

**Response Success (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "found": 150,
  "processed": 150,
  "progress": 100,
  "error": "",
  "created_at": "2024-11-02T10:30:00Z",
  "updated_at": "2024-11-02T10:35:00Z"
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "scan_id is required"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "scan not found"
}
```

**Códigos de estado:**
- `200`: Scan encontrado
- `400`: ID inválido o faltante
- `404`: Scan no encontrado

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8081/v1/scans/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: dev-secret-key"
```

---

## Files

### GET /v1/files

**Descripción**: Lista archivos con paginación y filtros opcionales por búsqueda, tipo y categoría

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/files`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)
  - `q` (string, optional): Búsqueda por nombre de archivo (similarity search)
  - `type` (string, optional): Filtrar por tipo de archivo (stl, zip, rar)
  - `category` (string, optional): Filtrar por nombre de categoría

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "path": "E:\\Impresion3D\\models\\dragon.stl",
      "file_name": "dragon.stl",
      "type": "stl",
      "size": 2048576,
      "modified_at": "2024-10-15T08:20:00Z",
      "sha256": "abc123...",
      "created_at": "2024-11-02T10:30:00Z",
      "updated_at": "2024-11-02T10:30:00Z",
      "categories": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "name": "miniatures",
          "created_at": "2024-11-01T00:00:00Z"
        }
      ]
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 20
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `500`: Error al listar archivos

**Ejemplo con cURL:**
```bash
# Listar todos los archivos
curl -X GET "http://localhost:8081/v1/files?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"

# Buscar archivos
curl -X GET "http://localhost:8081/v1/files?q=dragon" \
  -H "X-API-Key: dev-secret-key"

# Filtrar por tipo
curl -X GET "http://localhost:8081/v1/files?type=stl" \
  -H "X-API-Key: dev-secret-key"

# Filtrar por categoría
curl -X GET "http://localhost:8081/v1/files?category=miniatures" \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/files/{id}

**Descripción**: Obtiene el detalle de un archivo específico por su ID

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/files/{id}`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del archivo

**Response Success (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "path": "E:\\Impresion3D\\models\\dragon.stl",
  "file_name": "dragon.stl",
  "type": "stl",
  "size": 2048576,
  "modified_at": "2024-10-15T08:20:00Z",
  "sha256": "abc123...",
  "created_at": "2024-11-02T10:30:00Z",
  "updated_at": "2024-11-02T10:30:00Z",
  "categories": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "miniatures",
      "created_at": "2024-11-01T00:00:00Z"
    }
  ]
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "file_id is required"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "file not found"
}
```

**Códigos de estado:**
- `200`: Archivo encontrado
- `400`: ID inválido o faltante
- `404`: Archivo no encontrado

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8081/v1/files/660e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: dev-secret-key"
```

---

### POST /v1/files/{id}/reclassify

**Descripción**: Reclasifica un archivo usando OpenAI. Las categorías existentes se reemplazan por las nuevas sugeridas por la IA.

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: POST
- **URL**: `/v1/files/{id}/reclassify`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del archivo

**Response Success (200 OK):**
```json
{
  "file_id": "660e8400-e29b-41d4-a716-446655440001",
  "categories": ["miniatures", "fantasy"]
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "file_id is required"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "file not found"
}
```

**Response Error (503 Service Unavailable):**
```json
{
  "error": "OpenAI classification is not enabled"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": "classification failed"
}
```

**Códigos de estado:**
- `200`: Archivo reclasificado exitosamente
- `400`: ID inválido o faltante
- `404`: Archivo no encontrado
- `503`: OpenAI no habilitado (OPENAI_API_KEY no configurado)
- `500`: Error en clasificación

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8081/v1/files/660e8400-e29b-41d4-a716-446655440001/reclassify \
  -H "X-API-Key: dev-secret-key"
```

---

### PATCH /v1/files/{id}/categories

**Descripción**: Actualiza manualmente las categorías asignadas a un archivo (reemplaza las existentes)

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: PATCH
- **URL**: `/v1/files/{id}/categories`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del archivo
- **Body**:
  ```json
  {
    "category_ids": [
      "770e8400-e29b-41d4-a716-446655440002",
      "880e8400-e29b-41d4-a716-446655440003"
    ]
  }
  ```

**Validaciones:**
- `category_ids`: array de UUIDs válidos

**Response Success (200 OK):**
```json
{
  "file_id": "660e8400-e29b-41d4-a716-446655440001",
  "categories": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "miniatures",
      "created_at": "2024-11-01T00:00:00Z"
    },
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "name": "fantasy",
      "created_at": "2024-11-01T00:00:00Z"
    }
  ]
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "invalid request body"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "file not found"
}
```

**Códigos de estado:**
- `200`: Categorías actualizadas exitosamente
- `400`: Request inválido
- `404`: Archivo no encontrado
- `500`: Error al actualizar categorías

**Ejemplo con cURL:**
```bash
curl -X PATCH http://localhost:8081/v1/files/660e8400-e29b-41d4-a716-446655440001/categories \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-secret-key" \
  -d '{
    "category_ids": [
      "770e8400-e29b-41d4-a716-446655440002"
    ]
  }'
```

---

## Categories

### GET /v1/categories

**Descripción**: Lista todas las categorías disponibles con paginación y búsqueda. Solo retorna categorías activas (no eliminadas).

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/categories`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `q` (string, optional): Búsqueda por nombre de categoría (case-insensitive, usa ILIKE)
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "miniatures",
      "created_at": "2024-11-01T00:00:00Z",
      "deleted_at": null
    },
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "name": "fantasy",
      "created_at": "2024-11-01T00:00:00Z",
      "deleted_at": null
    }
  ],
  "total": 15,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `500`: Error al listar categorías

**Ejemplo con cURL:**
```bash
# Listar todas las categorías
curl -X GET "http://localhost:8081/v1/categories?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"

# Buscar categorías
curl -X GET "http://localhost:8081/v1/categories?q=lego" \
  -H "X-API-Key: dev-secret-key"
```

---

### POST /v1/categories

**Descripción**: Crea una nueva categoría

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: POST
- **URL**: `/v1/categories`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Body**:
  ```json
  {
    "name": "lego"
  }
  ```

**Validaciones:**
- `name`: string requerido, único (case-insensitive)

**Response Success (201 Created):**
```json
{
  "id": "990e8400-e29b-41d4-a716-446655440008",
  "name": "lego",
  "created_at": "2024-11-02T15:30:00Z",
  "deleted_at": null
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "name is required"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": "failed to create category"
}
```

**Códigos de estado:**
- `201`: Categoría creada exitosamente
- `400`: Request inválido
- `500`: Error al crear categoría (ej: nombre duplicado)

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8081/v1/categories \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-secret-key" \
  -d '{"name": "lego"}'
```

---

### GET /v1/categories/{id}

**Descripción**: Obtiene una categoría específica por su ID. Solo retorna categorías activas (no eliminadas).

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/categories/{id}`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID de la categoría

**Response Success (200 OK):**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "miniatures",
  "created_at": "2024-11-01T00:00:00Z",
  "deleted_at": null
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "Invalid category ID"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "category not found"
}
```

**Códigos de estado:**
- `200`: Categoría encontrada
- `400`: ID inválido
- `404`: Categoría no encontrada o eliminada
- `500`: Error al obtener categoría

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8081/v1/categories/770e8400-e29b-41d4-a716-446655440002 \
  -H "X-API-Key: dev-secret-key"
```

---

### PUT /v1/categories/{id}

**Descripción**: Actualiza el nombre de una categoría existente

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: PUT
- **URL**: `/v1/categories/{id}`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID de la categoría
- **Body**:
  ```json
  {
    "name": "lego-sets"
  }
  ```

**Validaciones:**
- `name`: string requerido, único (case-insensitive)

**Response Success (200 OK):**
```json
{
  "id": "990e8400-e29b-41d4-a716-446655440008",
  "name": "lego-sets",
  "created_at": "2024-11-02T15:30:00Z",
  "deleted_at": null
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "name is required"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "category not found"
}
```

**Códigos de estado:**
- `200`: Categoría actualizada exitosamente
- `400`: Request inválido o ID inválido
- `404`: Categoría no encontrada
- `500`: Error al actualizar categoría

**Notas:**
- Solo se puede actualizar el nombre de categorías activas (no eliminadas)
- Los cambios se reflejan automáticamente en todos los archivos y folders que usan esta categoría

**Ejemplo con cURL:**
```bash
curl -X PUT http://localhost:8081/v1/categories/990e8400-e29b-41d4-a716-446655440008 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-secret-key" \
  -d '{"name": "lego-sets"}'
```

---

### DELETE /v1/categories/{id}

**Descripción**: Elimina una categoría mediante soft delete (marca como eliminada sin borrar de la base de datos). Las relaciones con archivos y folders se mantienen intactas.

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: DELETE
- **URL**: `/v1/categories/{id}`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID de la categoría

**Response Success (200 OK):**
```json
{
  "message": "category deleted successfully"
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "Invalid category ID"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": "failed to delete category"
}
```

**Códigos de estado:**
- `200`: Categoría eliminada exitosamente
- `400`: ID inválido
- `500`: Error al eliminar categoría

**Notas:**
- Soft delete: La categoría no se elimina físicamente, solo se marca como eliminada (deleted_at)
- Las categorías eliminadas no aparecen en listados ni búsquedas
- Las relaciones con archivos/folders se mantienen
- Una categoría eliminada puede restaurarse con POST /v1/categories/{id}/restore

**Ejemplo con cURL:**
```bash
curl -X DELETE http://localhost:8081/v1/categories/990e8400-e29b-41d4-a716-446655440008 \
  -H "X-API-Key: dev-secret-key"
```

---

### POST /v1/categories/{id}/restore

**Descripción**: Restaura una categoría previamente eliminada (soft delete), haciéndola visible nuevamente

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: POST
- **URL**: `/v1/categories/{id}/restore`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID de la categoría

**Response Success (200 OK):**
```json
{
  "message": "category restored successfully"
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "Invalid category ID"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": "failed to restore category"
}
```

**Códigos de estado:**
- `200`: Categoría restaurada exitosamente
- `400`: ID inválido
- `500`: Error al restaurar categoría

**Notas:**
- Solo se pueden restaurar categorías que fueron previamente eliminadas con soft delete
- Después de restaurar, la categoría vuelve a aparecer en listados y búsquedas
- Las relaciones con archivos/folders se mantuvieron intactas durante la eliminación

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8081/v1/categories/990e8400-e29b-41d4-a716-446655440008/restore \
  -H "X-API-Key: dev-secret-key"
```

---

## Browse & Navigation

### GET /v1/browse

**Descripción**: Lista solo los folders raíz (sin parent_folder_id) con paginación y búsqueda. No incluye archivos raíz. Útil para navegación de folders.

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/browse`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `q` (string, optional): Búsqueda por nombre de folder (case-insensitive, usa ILIKE)
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "name": "Miniatures",
      "type": "folder",
      "file_count": 45,
      "categories": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "name": "miniatures",
          "created_at": "2024-11-01T00:00:00Z"
        }
      ],
      "created_at": "2024-11-02T10:30:00Z"
    }
  ],
  "total": 12,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `500`: Error al listar folders

**Ejemplo con cURL:**
```bash
# Listar todos los folders raíz
curl -X GET "http://localhost:8081/v1/browse?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"

# Buscar folders por nombre
curl -X GET "http://localhost:8081/v1/browse?q=miniatures&page=1" \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/mixed

**Descripción**: Vista mixta tipo explorador de archivos. Sin `folder_id` muestra folders raíz + archivos raíz. Con `folder_id` muestra subfolders + archivos de ese folder.

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/mixed`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)
  - `folder_id` (string, optional): UUID del folder para mostrar su contenido

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "name": "Miniatures",
      "type": "folder",
      "file_count": 45,
      "categories": [],
      "created_at": "2024-11-02T10:30:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "dragon.stl",
      "type": "stl",
      "size": 2048576,
      "categories": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "name": "miniatures",
          "created_at": "2024-11-01T00:00:00Z"
        }
      ],
      "created_at": "2024-11-02T10:30:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "page_size": 20,
  "total_pages": 2
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `400`: folder_id inválido
- `500`: Error al listar contenido

**Ejemplo con cURL:**
```bash
# Vista raíz (folders + archivos raíz)
curl -X GET "http://localhost:8081/v1/mixed?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"

# Contenido de un folder
curl -X GET "http://localhost:8081/v1/mixed?folder_id=990e8400-e29b-41d4-a716-446655440004" \
  -H "X-API-Key: dev-secret-key"
```

---

## Folders

### GET /v1/folders

**Descripción**: Lista todos los folders con paginación y búsqueda, incluyendo file count y categorías

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/folders`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **Query Params**:
  - `q` (string, optional): Búsqueda por nombre de folder (case-insensitive, usa ILIKE)
  - `page` (number, optional): Número de página (default: 1)
  - `page_size` (number, optional): Elementos por página (default: 20, max: 100)

**Response Success (200 OK):**
```json
{
  "items": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "name": "Miniatures",
      "path": "E:\\Impresion3D\\Miniatures",
      "parent_folder_id": null,
      "created_at": "2024-11-02T10:30:00Z",
      "updated_at": "2024-11-02T10:30:00Z",
      "file_count": 45,
      "categories": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "name": "miniatures",
          "created_at": "2024-11-01T00:00:00Z"
        }
      ]
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `500`: Error al listar folders

**Ejemplo con cURL:**
```bash
# Listar todos los folders
curl -X GET "http://localhost:8081/v1/folders?page=1&page_size=20" \
  -H "X-API-Key: dev-secret-key"

# Buscar folders por nombre
curl -X GET "http://localhost:8081/v1/folders?q=warhammer&page=1" \
  -H "X-API-Key: dev-secret-key"
```

---

### GET /v1/folders/{id}

**Descripción**: Obtiene un folder específico con sus subfolders y archivos. Los archivos están paginados para manejar folders con muchos archivos (ej: 1000+).

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: GET
- **URL**: `/v1/folders/{id}`
- **Headers**:
  ```json
  {
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del folder
- **Query Params**:
  - `page` (number, optional): Número de página para archivos (default: 1)
  - `page_size` (number, optional): Archivos por página (default: 50, max: 100)
  - `search` (string, optional): Búsqueda por nombre (aplica a subfolders y archivos)
  - `type` (string, optional): Filtrar archivos por tipo (stl, zip, rar)
  - `category` (string, optional): Filtrar por nombre de categoría (aplica a subfolders y archivos)

**Response Success (200 OK):**
```json
{
  "folder": {
    "id": "990e8400-e29b-41d4-a716-446655440004",
    "name": "Miniatures",
    "path": "E:\\Impresion3D\\Miniatures",
    "parent_folder_id": null,
    "created_at": "2024-11-02T10:30:00Z",
    "updated_at": "2024-11-02T10:30:00Z"
  },
  "subfolders": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440005",
      "name": "Fantasy",
      "path": "E:\\Impresion3D\\Miniatures\\Fantasy",
      "parent_folder_id": "990e8400-e29b-41d4-a716-446655440004",
      "created_at": "2024-11-02T10:30:00Z",
      "updated_at": "2024-11-02T10:30:00Z",
      "file_count": 20,
      "categories": []
    }
  ],
  "files": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "path": "E:\\Impresion3D\\Miniatures\\dragon.stl",
      "file_name": "dragon.stl",
      "type": "stl",
      "size": 2048576,
      "modified_at": "2024-10-15T08:20:00Z",
      "sha256": "abc123...",
      "created_at": "2024-11-02T10:30:00Z",
      "updated_at": "2024-11-02T10:30:00Z",
      "categories": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "name": "miniatures",
          "created_at": "2024-11-01T00:00:00Z"
        }
      ]
    }
  ],
  "categories": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "miniatures",
      "created_at": "2024-11-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 1250,
    "page": 1,
    "page_size": 50,
    "total_pages": 25
  }
}
```

**Notas sobre paginación:**
- `subfolders`: Se retornan completos (sin paginar). Raramente hay cientos de subfolders.
- `files`: Paginados según `page` y `page_size`. Esto resuelve el problema de folders con 1000+ archivos.
- Si hay filtros activos (`search`, `type`, `category`): se aplican primero y luego se pagina el resultado filtrado.
- Sin filtros: la paginación es eficiente a nivel de base de datos.

**Response Error (400 Bad Request):**
```json
{
  "error": "Invalid folder ID"
}
```

**Response Error (404 Not Found):**
```json
{
  "error": "Folder not found"
}
```

**Códigos de estado:**
- `200`: Folder encontrado
- `400`: ID inválido
- `404`: Folder no encontrado
- `500`: Error al obtener folder

**Ejemplo con cURL:**
```bash
# Obtener folder con primera página de archivos (50 archivos)
curl -X GET http://localhost:8081/v1/folders/990e8400-e29b-41d4-a716-446655440004 \
  -H "X-API-Key: dev-secret-key"

# Obtener página 2 con 100 archivos por página
curl -X GET "http://localhost:8081/v1/folders/990e8400-e29b-41d4-a716-446655440004?page=2&page_size=100" \
  -H "X-API-Key: dev-secret-key"

# Con filtros (paginación se aplica al resultado filtrado)
curl -X GET "http://localhost:8081/v1/folders/990e8400-e29b-41d4-a716-446655440004?search=dragon&type=stl&page=1" \
  -H "X-API-Key: dev-secret-key"
```

---

### PATCH /v1/folders/{id}/categories

**Descripción**: Actualiza las categorías de un folder con opciones de propagación a archivos y subfolders. Usa batch operations para máxima eficiencia (optimizado para folders con 1000+ archivos).

**Autenticación**: Sí (X-API-Key)

**Request:**
- **Method**: PATCH
- **URL**: `/v1/folders/{id}/categories`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "X-API-Key": "dev-secret-key"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del folder
- **Body**:
  ```json
  {
    "category_ids": [
      "770e8400-e29b-41d4-a716-446655440002"
    ],
    "apply_to_stl": true,
    "apply_to_zip": false,
    "apply_to_rar": false,
    "apply_to_subfolders": true
  }
  ```

**Validaciones:**
- `category_ids`: array de UUIDs válidos
- `apply_to_stl`: (boolean, optional) Aplicar a archivos .stl del folder
- `apply_to_zip`: (boolean, optional) Aplicar a archivos .zip del folder
- `apply_to_rar`: (boolean, optional) Aplicar a archivos .rar del folder
- `apply_to_subfolders`: (boolean, optional) Aplicar recursivamente a subfolders

**Response Success (200 OK):**
```json
{
  "categories": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "miniatures",
      "created_at": "2024-11-01T00:00:00Z"
    }
  ]
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": "Invalid folder ID"
}
```

**Códigos de estado:**
- `200`: Categorías actualizadas exitosamente
- `400`: Request inválido
- `500`: Error al actualizar categorías

**Notas:**
- La propagación reemplaza las categorías existentes en archivos/subfolders
- La propagación a subfolders es recursiva (afecta a todos los descendientes)

**Ejemplo con cURL:**
```bash
curl -X PATCH http://localhost:8081/v1/folders/990e8400-e29b-41d4-a716-446655440004/categories \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-secret-key" \
  -d '{
    "category_ids": ["770e8400-e29b-41d4-a716-446655440002"],
    "apply_to_stl": true,
    "apply_to_zip": false,
    "apply_to_rar": false,
    "apply_to_subfolders": true
  }'
```

---

## Convenciones Generales

### Autenticación

Todos los endpoints requieren el header `X-API-Key`:

```
X-API-Key: dev-secret-key
```

El valor se configura en la variable de entorno `API_KEY`.

---

### Formatos de Fecha

Todas las fechas están en formato ISO 8601 con timezone UTC:
```
2024-11-02T10:30:00Z
```

---

### Paginación

Los endpoints que retornan listas incluyen paginación con los siguientes query params:

- `page`: Número de página (default: 1, mínimo: 1)
- `page_size`: Elementos por página (default: 20, máximo: 100)

La respuesta incluye metadatos:

```json
{
  "items": [...],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "total_pages": 8
}
```

---

### Manejo de Errores

Errores retornan un objeto simple:

```json
{
  "error": "mensaje de error descriptivo"
}
```

**Códigos HTTP comunes:**
- `200 OK`: Operación exitosa
- `201 Created`: Recurso creado exitosamente
- `202 Accepted`: Operación aceptada (procesamiento asíncrono)
- `400 Bad Request`: Request inválido (parámetros faltantes o inválidos)
- `401 Unauthorized`: API Key faltante o inválida
- `404 Not Found`: Recurso no encontrado
- `500 Internal Server Error`: Error interno del servidor
- `503 Service Unavailable`: Servicio no disponible (ej: OpenAI deshabilitado)

---

### CORS

Dominios permitidos:
- `*` (todos los orígenes en desarrollo)

Headers permitidos:
- `Accept`
- `Authorization`
- `Content-Type`
- `X-API-Key`

---

### Variables de Entorno Importantes

```env
# Server
PORT=8081

# Database
DATABASE_URL=postgresql://...

# OpenAI (opcional)
OPENAI_API_KEY=sk-...

# Scan
SCAN_ROOT_DIR=E:\Impresion3D
SUPPORTED_EXTS=.stl,.zip,.rar

# Security
API_KEY=dev-secret-key
```

---

### Tipos de Archivo Soportados

- `.stl` - Archivos de modelos 3D
- `.zip` - Archivos comprimidos
- `.rar` - Archivos comprimidos

---

### Categorías

Las categorías se crean automáticamente durante el primer scan desde un conjunto predefinido en la base de datos.

Categorías comunes:
- `miniatures`
- `fantasy`
- `sci-fi`
- `tools`
- `uncategorized` (default si OpenAI no clasifica)

---

**Mantenimiento**: Este documento debe actualizarse cada vez que se cree, modifique o elimine un endpoint.
