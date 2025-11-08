# API Documentation

**Base URL**: `http://localhost:8080/api` (development)

**Última actualización**: 2024-11-02

---

## Índice de Endpoints

- [Users](#users)
  - [POST /api/users](#post-apiusers) - Crear usuario
  - [GET /api/users/:id](#get-apiusersid) - Obtener usuario por ID
  - [GET /api/users](#get-apiusers) - Listar usuarios

---

## Users

### POST /api/users

**Descripción**: Crea un nuevo usuario en el sistema

**Autenticación**: No requerida

**Request:**

- **Method**: POST
- **URL**: `/api/users`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json"
  }
  ```
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securePassword123",
    "name": "John Doe"
  }
  ```

**Validaciones:**
- `email`: requerido, formato de email válido, único en el sistema
- `password`: requerido, mínimo 8 caracteres
- `name`: requerido, máximo 100 caracteres

**Response Success (201 Created):**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-11-02T10:30:00Z"
  },
  "message": "User created successfully"
}
```

**Response Error (400 Bad Request):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      "email is required",
      "password must be at least 8 characters"
    ]
  }
}
```

**Response Error (409 Conflict):**
```json
{
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email already exists",
    "details": []
  }
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "details": []
  }
}
```

**Códigos de estado:**
- `201`: Usuario creado exitosamente
- `400`: Datos de entrada inválidos
- `409`: Email ya existe en el sistema
- `500`: Error interno del servidor

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123",
    "name": "John Doe"
  }'
```

---

### GET /api/users/:id

**Descripción**: Obtiene la información de un usuario específico por su ID

**Autenticación**: Sí (Bearer Token)

**Request:**

- **Method**: GET
- **URL**: `/api/users/:id`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer <your-jwt-token>"
  }
  ```
- **URL Params**:
  - `id` (string, required): UUID del usuario

**Response Success (200 OK):**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-11-02T10:30:00Z"
  }
}
```

**Response Error (401 Unauthorized):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid authentication token",
    "details": []
  }
}
```

**Response Error (404 Not Found):**
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found",
    "details": []
  }
}
```

**Códigos de estado:**
- `200`: Usuario encontrado
- `401`: No autenticado o token inválido
- `404`: Usuario no encontrado
- `500`: Error interno del servidor

**Ejemplo con cURL:**
```bash
curl -X GET http://localhost:8080/api/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer your-jwt-token"
```

---

### GET /api/users

**Descripción**: Lista todos los usuarios con paginación

**Autenticación**: Sí (Bearer Token)

**Request:**

- **Method**: GET
- **URL**: `/api/users`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer <your-jwt-token>"
  }
  ```
- **Query Params**:
  - `page` (number, optional): Número de página (default: 1)
  - `limit` (number, optional): Usuarios por página (default: 10, max: 100)
  - `search` (string, optional): Buscar por nombre o email

**Response Success (200 OK):**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user1@example.com",
      "name": "John Doe",
      "created_at": "2024-11-02T10:30:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "email": "user2@example.com",
      "name": "Jane Smith",
      "created_at": "2024-11-02T11:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

**Response Error (401 Unauthorized):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid authentication token",
    "details": []
  }
}
```

**Códigos de estado:**
- `200`: Lista obtenida exitosamente
- `401`: No autenticado o token inválido
- `500`: Error interno del servidor

**Ejemplo con cURL:**
```bash
curl -X GET "http://localhost:8080/api/users?page=1&limit=10" \
  -H "Authorization: Bearer your-jwt-token"
```

---

## Plantilla para nuevos endpoints

Copia esta plantilla cuando documentes un nuevo endpoint:

```markdown
### [MÉTODO] /api/[ruta]

**Descripción**: [Descripción breve de qué hace el endpoint]

**Autenticación**: [Sí/No] [(tipo de autenticación si aplica)]

**Request:**

- **Method**: [GET/POST/PUT/PATCH/DELETE]
- **URL**: `/api/[ruta]`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <token>" // si requiere auth
  }
  ```
- **URL Params** (si aplica):
  - `param_name` (type, required/optional): Descripción
  
- **Query Params** (si aplica):
  - `param_name` (type, required/optional): Descripción
  
- **Body** (si aplica):
  ```json
  {
    "field1": "value",
    "field2": 123
  }
  ```

**Validaciones** (si aplica):
- `field1`: reglas de validación
- `field2`: reglas de validación

**Response Success ([código]):**
```json
{
  // estructura de respuesta exitosa
}
```

**Response Error ([código]):**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": []
  }
}
```

**Códigos de estado:**
- `xxx`: Descripción
- `xxx`: Descripción

**Ejemplo con cURL:**
```bash
curl -X [METHOD] http://localhost:8080/api/[ruta] \
  -H "Content-Type: application/json" \
  [opciones adicionales]
```

---
```

---

## Notas importantes

### Formatos de fecha
Todas las fechas están en formato ISO 8601: `2024-11-02T10:30:00Z`

### Paginación
Los endpoints que retornan listas incluyen metadatos de paginación:
```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

### Rate Limiting
- Límite: 100 requests por minuto por IP
- Headers de respuesta incluyen:
  - `X-RateLimit-Limit`: Límite total
  - `X-RateLimit-Remaining`: Requests restantes
  - `X-RateLimit-Reset`: Timestamp cuando se resetea

### CORS
Dominios permitidos en desarrollo:
- `http://localhost:5173` (Vite default)
- `http://localhost:3000`

---

## Convenciones

### Códigos de error comunes
- `VALIDATION_ERROR`: Error de validación de datos de entrada
- `UNAUTHORIZED`: Token faltante o inválido
- `FORBIDDEN`: Sin permisos para esta operación
- `NOT_FOUND`: Recurso no encontrado
- `CONFLICT`: Conflicto con estado actual (ej: email duplicado)
- `INTERNAL_ERROR`: Error interno del servidor

### Naming
- Endpoints: snake_case o kebab-case según preferencia
- JSON fields: snake_case
- HTTP methods: Seguir semántica REST

---

## Autenticación

### Obtener token
```bash
POST /api/auth/login
Body: { "email": "user@example.com", "password": "password" }
Response: { "token": "jwt-token-here" }
```

### Usar token
Incluir en header de todas las requests autenticadas:
```
Authorization: Bearer <tu-token-jwt>
```

### Renovar token
```bash
POST /api/auth/refresh
Header: Authorization: Bearer <tu-refresh-token>
Response: { "token": "nuevo-jwt-token" }
```

---

**Mantenimiento**: Este documento debe actualizarse cada vez que se cree, modifique o elimine un endpoint.
