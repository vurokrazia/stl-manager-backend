# Backend API - Go

## Agente activo

**IMPORTANTE**: Antes de responder a CUALQUIER solicitud, SIEMPRE debes:
1. Leer completamente el archivo: `.claude/agents/go-backend-expert.md`
2. Seguir TODAS las reglas definidas en ese archivo
3. Aplicar el protocolo anti-suposiciones obligatoriamente

Si no has leído ese archivo en esta sesión, léelo AHORA antes de continuar.

---

## Stack tecnológico

### Lenguaje y framework
- Go 1.21+
- Framework: [COMPLETA: Gin/Echo/Fiber/Chi/net-http]
- Router: [COMPLETA si aplica]

### Base de datos
- Database: [COMPLETA: PostgreSQL/MySQL/MongoDB/SQLite]
- ORM/Driver: [COMPLETA: GORM/sqlx/pgx/queries raw]

### Autenticación
- JWT: [Sí/No]
- Librería: [COMPLETA: golang-jwt/jwt o similar]

### Otras librerías
- Validación: [COMPLETA: go-playground/validator o similar]
- Logging: [COMPLETA: log standard/logrus/zap]
- Variables de entorno: [COMPLETA: godotenv/viper]

---

## Estructura del proyecto

```
backend/
├── cmd/
│   └── api/
│       └── main.go           - Entry point
├── internal/
│   ├── handlers/             - HTTP handlers (controllers)
│   ├── services/             - Lógica de negocio
│   ├── repository/           - Acceso a datos
│   ├── models/               - Structs y tipos
│   └── middleware/           - Middlewares (auth, logging, etc)
├── docs/
│   └── api/
│       └── endpoints.md      - DOCUMENTACIÓN DE ENDPOINTS
├── .env                      - Variables de entorno (NO commitear)
├── go.mod
└── go.sum
```

---

## Convenciones de código

### Naming
- Archivos: `snake_case.go` (ej: `user_handler.go`)
- Funciones exported: `PascalCase` (ej: `CreateUser`)
- Funciones unexported: `camelCase` (ej: `validateEmail`)
- Constantes: `PascalCase` o `UPPER_SNAKE_CASE` según contexto

### Estructura de handlers
```go
// internal/handlers/user_handler.go
func CreateUser(c *gin.Context) {
    // 1. Validar input
    // 2. Llamar service
    // 3. Retornar response
}
```

### Estructura de services
```go
// internal/services/user_service.go
type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(data CreateUserDTO) (*User, error) {
    // Lógica de negocio aquí
}
```

---

## Formato de respuestas API

### Success
```go
type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}
```

### Error
```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string   `json:"code"`
    Message string   `json:"message"`
    Details []string `json:"details,omitempty"`
}
```

---

## Códigos de error estándar

- `VALIDATION_ERROR` - Error de validación
- `UNAUTHORIZED` - No autenticado
- `FORBIDDEN` - Sin permisos
- `NOT_FOUND` - Recurso no encontrado
- `CONFLICT` - Conflicto (ej: email duplicado)
- `INTERNAL_ERROR` - Error interno

---

## Documentación de endpoints

**IMPORTANTE**: 
- CADA endpoint nuevo debe documentarse en `docs/api/endpoints.md`
- Usar el formato de plantilla que está en ese archivo
- Incluir: método, ruta, body, validaciones, responses, códigos de estado, ejemplo cURL

---

## Variables de entorno

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/dbname

# JWT (si aplica)
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

---

## Comandos útiles

```bash
# Desarrollo
go run cmd/api/main.go

# Build
go build -o bin/api cmd/api/main.go

# Tests
go test ./...

# Linting
golangci-lint run

# Hot reload (si usas air)
air
```

---

## Testing

- Unit tests: `*_test.go` al lado del archivo
- Cobertura mínima: [COMPLETA: ej 80%]
- Naming: `TestFunctionName`

---

## Reglas importantes

1. **NUNCA hagas suposiciones** - Si hay duda, pregunta
2. **SIEMPRE documenta endpoints** en `docs/api/endpoints.md`
3. **SIEMPRE valida inputs** del usuario
4. **SIEMPRE maneja errores** explícitamente
5. **NUNCA commitees** archivos `.env` o secrets
6. **SIEMPRE usa** error wrapping con `fmt.Errorf("context: %w", err)`

---

## Workflow de desarrollo

1. Recibir tarea
2. Hacer preguntas de clarificación
3. Confirmar entendimiento
4. ESPERAR confirmación
5. Implementar código
6. Documentar en `docs/api/endpoints.md`
7. Reportar archivos creados/modificados

---

## Notas del proyecto

[AGREGA AQUÍ NOTAS ESPECÍFICAS DE TU PROYECTO]

- Feature actual en desarrollo: 
- Endpoints completados: 
- Endpoints pendientes: 
- Issues conocidos: 

---

**Última actualización**: 2024-11-02
