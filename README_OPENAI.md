# OpenAI Classification

Este backend incluye clasificación automática de archivos usando OpenAI GPT-4o-mini.

## Cómo Funciona

### ✅ CON API Key (Clasificación Habilitada)

Si configuras `OPENAI_API_KEY` en el archivo `.env`:

1. **Durante el Scan**:
   - Cada archivo se clasifica automáticamente
   - Se asignan 0-3 categorías por archivo
   - Si no hay coincidencias, se marca como "uncategorized"

2. **Endpoint Reclassify**:
   - `/v1/files/{id}/reclassify` funciona normalmente
   - Vuelve a clasificar el archivo con OpenAI

3. **¿Qué se envía a OpenAI?**
   - Nombre del archivo (ej: "porsche_911_turbo.stl")
   - Lista de categorías permitidas
   - NO se envían rutas completas ni contenido de archivos

### ❌ SIN API Key (Clasificación Desactivada)

Si dejas `OPENAI_API_KEY` vacío en `.env`:

1. **Durante el Scan**:
   - Archivos se guardan normalmente en la base de datos
   - Todos se marcan como "uncategorized"
   - NO se hace ninguna llamada a OpenAI

2. **Endpoint Reclassify**:
   - `/v1/files/{id}/reclassify` devuelve error 503
   - Mensaje: "OpenAI classification is not enabled"

3. **Sin Logs en OpenAI**:
   - Nada se registra en tu cuenta de OpenAI
   - Nadie puede ver qué archivos tienes

## Configuración

```bash
# En .env - Para ACTIVAR clasificación:
OPENAI_API_KEY=sk-proj-xxxxxxxxxxxxxxxxxxxxx

# Para DESACTIVAR clasificación (dejar vacío):
OPENAI_API_KEY=
```

## Privacidad

### Lo que OpenAI NO recibe:
- ❌ Rutas completas de archivos
- ❌ Contenido de los archivos
- ❌ Información de tu sistema
- ❌ Credenciales de base de datos

### Lo que OpenAI SÍ recibe (solo si está habilitado):
- ✅ Nombre del archivo (ej: "porsche_911.stl")
- ✅ Lista de categorías disponibles
- ✅ Timestamp de la petición

### Logs en OpenAI Dashboard
Si alguien tiene acceso a tu cuenta de platform.openai.com puede ver:
- Los nombres de archivos que fueron clasificados
- Las categorías asignadas
- Tokens consumidos y costo

**Solución**: No compartas tu cuenta de OpenAI o deja el API key vacío.

## Recomendación

Si te preocupa la privacidad:
1. Deja `OPENAI_API_KEY` vacío
2. Clasifica manualmente desde el frontend
3. O usa OpenAI solo temporalmente para clasificar y luego elimina el key

El sistema funciona perfectamente sin OpenAI, solo no tendrás clasificación automática.
