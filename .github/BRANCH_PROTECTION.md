# Configuración de Branch Protection Rules

Este documento explica cómo configurar las reglas de protección de ramas en GitHub para prevenir merges sin que pasen todos los tests y validaciones.

## Paso 1: Acceder a Branch Protection Rules

1. Ve a tu repositorio en GitHub: `https://github.com/vurokrazia/stl-manager-backend`
2. Click en **Settings** (Configuración)
3. En el menú lateral, click en **Branches** (bajo "Code and automation")
4. En "Branch protection rules", click en **Add rule** o edita la regla existente para `master`

## Paso 2: Configurar la Regla para `master`

### Branch name pattern
```
master
```

### Configuraciones Requeridas

Marca las siguientes opciones:

#### ✅ Require a pull request before merging
- **Require approvals**: 1 (opcional, ajusta según tus necesidades)
- **Dismiss stale pull request approvals when new commits are pushed**: ✅

#### ✅ Require status checks to pass before merging
- **Require branches to be up to date before merging**: ✅

**Status checks que DEBEN pasar** (agrega estos):
- `Test & Build` - Job principal de tests
- `Code Formatting` - Verifica formato del código
- `Security Scan` - Escaneo de seguridad (opcional, pero recomendado)

Para agregar estos checks:
1. Click en el campo de búsqueda "Search for status checks"
2. Escribe el nombre del check (aparecerán después del primer PR)
3. Selecciona cada uno

#### ✅ Require conversation resolution before merging
Asegura que todos los comentarios del PR estén resueltos.

#### ✅ Do not allow bypassing the above settings
Importante: ni siquiera los admins pueden hacer merge sin pasar los checks.

#### Opcionales pero Recomendados

- ✅ **Require linear history**: Evita merge commits
- ✅ **Require deployments to succeed before merging**: Si tienes deploys automáticos
- ✅ **Lock branch**: Solo lectura (solo si quieres que nadie pushee directamente)

### Configuración para permitir actualizaciones automáticas
- ❌ **Allow force pushes**: NO (previene reescribir historia)
- ❌ **Allow deletions**: NO (previene borrar la rama accidentalmente)

## Paso 3: Guardar Cambios

Click en **Create** o **Save changes** al final de la página.

## Paso 4: Verificar que Funciona

1. Crea un PR desde cualquier rama hacia `master`
2. Verás que GitHub automáticamente ejecuta los workflows
3. No podrás hacer merge hasta que todos los checks pasen

Ejemplo:

```
✅ Test & Build — Required
✅ Code Formatting — Required
✅ Security Scan — Required

[ Merge pull request ] ← Este botón estará deshabilitado si algún check falla
```

## Configuración Actual del Workflow

El workflow `.github/workflows/test.yml` ejecuta automáticamente:

### Job: Test & Build (CRÍTICO)
- ✅ Instala dependencias
- ✅ Verifica dependencias con `go mod verify`
- ✅ Setup de PostgreSQL con extensiones
- ✅ Ejecuta migraciones
- ✅ **Linting** con golangci-lint (FALLA si hay errores)
- ✅ **Tests de integración** con coverage (FALLA si hay tests fallidos)
- ✅ **Build** del binario (FALLA si no compila)
- ✅ **Unit tests** (FALLA si hay tests fallidos)

### Job: Security Scan (RECOMENDADO)
- Escaneo de seguridad con Gosec
- Detecta vulnerabilidades comunes

### Job: Code Formatting (CRÍTICO)
- Verifica que todo el código esté formateado con `gofmt`
- FALLA si hay archivos sin formatear

## Triggers Actuales

El workflow se ejecuta en:

```yaml
on:
  push:
    branches: [ main, master, develop, claude/**, test/** ]
  pull_request:
    branches: [ main, master, develop ]
```

## Resultados de Tests

Actualmente el proyecto tiene:

- **66 tests de integración** cubriendo todos los endpoints del API
- Cobertura de código con coverage reports
- Tests con race detector habilitado

### Módulos testeados:
- Categories API (17 tests)
- Files API (15 tests)
- Folders API (17 tests)
- Browse API (5 tests)
- Scans API (10 tests)
- Health API (2 tests)

## Troubleshooting

### "Status checks no aparecen en Branch Protection"

Los status checks solo aparecen DESPUÉS de que se ejecutan por primera vez.

**Solución**:
1. Crea un PR de prueba
2. Espera a que los workflows se ejecuten
3. Luego regresa a Branch Protection y los verás disponibles en el buscador

### "El merge está bloqueado pero los checks pasaron"

Verifica:
1. Que no haya conversaciones sin resolver (si activaste esa opción)
2. Que la rama esté actualizada con `master` (si activaste "require up to date")

### "Quiero hacer un hotfix de emergencia"

Si REALMENTE necesitas hacer merge sin pasar checks (NO RECOMENDADO):

1. Temporalmente deshabilita "Do not allow bypassing"
2. Haz el merge
3. **INMEDIATAMENTE vuelve a habilitar** la protección

**MEJOR OPCIÓN**: Crea un branch `hotfix/*` y:
1. Haz el fix
2. Espera a que pasen los tests (suelen tardar ~5-10 minutos)
3. Merge normalmente

## Comandos Útiles Locales

Antes de hacer push, ejecuta localmente para verificar:

```bash
# Run all tests
go test ./tests/integration/... -v

# Run linting
golangci-lint run --timeout=5m ./cmd/... ./internal/... ./tests/...

# Check formatting
gofmt -l .

# Fix formatting
gofmt -w .

# Build
go build -o bin/api ./cmd/api
```

---

## Resumen

Con esta configuración:

1. ✅ **Nadie puede hacer merge sin que pasen TODOS los tests**
2. ✅ **El código debe estar formateado correctamente**
3. ✅ **El linter debe pasar sin errores**
4. ✅ **La aplicación debe compilar correctamente**
5. ✅ **Security scan debe completarse**

Esto garantiza la calidad del código en `master` y previene bugs en producción.

---

**Última actualización**: 2024-11-09
