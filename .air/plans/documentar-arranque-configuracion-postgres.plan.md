## 1. Goal

Documentar con comentarios breves y claros el funcionamiento de los tres archivos solicitados, manteniendo intacto el comportamiento del backend.

## 2. Approach

Añadir comentarios en español únicamente en los puntos que representan una responsabilidad o decisión importante: carga de configuración, conexión de dependencias, composición de casos de uso, rutas HTTP, límites del servidor y apagado ordenado. Se conservarán los comentarios existentes cuando sean útiles y se evitará comentar cada línea para mantener el código legible.

## 3. File Changes

- **Modify** [main.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/cmd/api/main.go?type=file&root=%252F), especialmente líneas 26-263:
  - Explicar el arranque de la aplicación y la carga de configuración.
  - Documentar la inicialización y liberación de PostgreSQL y Redis.
  - Describir la inyección de dependencias para usuarios.
  - Explicar los middlewares, endpoints de salud y montaje de rutas.
  - Aclarar los timeouts del servidor, el manejo de errores y el apagado ordenado.

- **Modify** [config.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/config/config.go?type=file&root=%252F), líneas 12-74:
  - Describir los campos de configuración.
  - Explicar la carga opcional de .env, los valores por defecto y la precedencia del entorno.
  - Documentar las validaciones de conexiones.
  - Aclarar el comportamiento de getEnv y getEnvInt32, incluyendo el uso del valor alternativo cuando un entero no se puede convertir.

- **Modify** [postgres.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/database/postgres.go?type=file&root=%252F), líneas 11-49:
  - Explicar la configuración del pool.
  - Documentar los límites de conexiones y los tiempos de renovación, inactividad y comprobación.
  - Aclarar que la conexión se valida con Ping y que el pool se cierra si esa validación falla.

## 4. Implementation Steps

### Task 1: Documentar el punto de entrada

1. Añadir un comentario breve antes de main en [main.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/cmd/api/main.go?type=file&linesData=%7B%22range%22%3A%7B%22first%22%3A527%2C%22second%22%3A541%7D%2C%22lines%22%3A%7B%22first%22%3A25%2C%22second%22%3A26%7D%7D&root=%252F) indicando que inicia la aplicación y conecta sus dependencias.
2. Comentar la inicialización y limpieza de PostgreSQL y Redis en las secciones correspondientes.
3. Comentar la creación del repositorio, casos de uso y handler de usuarios como composición de capas.
4. Documentar el propósito de cada grupo de middlewares y de las rutas /health, /health/ready y /api/v1/users.
5. Explicar los límites de tiempo del servidor, la goroutine que lo inicia y la espera de errores o señales.
6. Mantener los comentarios existentes de salud y apagado, ajustándolos solo si hace falta para evitar duplicación.

### Task 2: Documentar la configuración

1. Añadir un comentario breve a Config en [config.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/config/config.go?type=file&linesData=%7B%22range%22%3A%7B%22first%22%3A196%2C%22second%22%3A367%7D%2C%22lines%22%3A%7B%22first%22%3A11%2C%22second%22%3A18%7D%7D&root=%252F) describiendo que reúne la configuración de la aplicación y sus servicios.
2. Explicar en Load que .env es opcional, que las variables del entorno se consultan con valores por defecto y que la URL actual de PostgreSQL está definida directamente en el código.
3. Documentar las validaciones de URL y límites del pool.
4. Añadir comentarios cortos a getEnv y getEnvInt32 describiendo sus valores de retorno y fallback.

### Task 3: Documentar la conexión PostgreSQL

1. Añadir un comentario breve a Config en [postgres.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/database/postgres.go?type=file&linesData=%7B%22range%22%3A%7B%22first%22%3A174%2C%22second%22%3A270%7D%2C%22lines%22%3A%7B%22first%22%3A10%2C%22second%22%3A14%7D%7D&root=%252F) indicando que contiene los parámetros del pool.
2. Documentar el flujo de NewPostgres: interpretar la URL, aplicar límites, crear el pool, comprobar conectividad y devolverlo.
3. Mantener el comentario existente sobre renovación y comprobación periódica, haciéndolo más específico si es necesario.
4. No modificar mensajes de error, tiempos, tipos ni comportamiento de cierre.

## 5. Acceptance Criteria

- Los tres archivos contienen comentarios en español que explican sus bloques principales sin comentar cada instrucción individual.
- main.go deja claro el flujo de arranque, dependencias, rutas, servidor y apagado.
- config.go deja claros los valores por defecto, la carga de .env, las validaciones y el comportamiento de los helpers.
- postgres.go deja claro cómo se construye, configura, comprueba y cierra el pool ante un fallo inicial.
- La URL de PostgreSQL actualmente definida en código queda documentada como comportamiento existente, sin introducir cambios funcionales.
- No se modifican rutas HTTP, valores por defecto, tiempos, mensajes de error ni firmas públicas.
- El código continúa compilando y los comentarios respetan el formato de Go.

## 6. Verification Steps

- Revisar los tres archivos para confirmar que los comentarios son breves, correctos y están junto al bloque que explican.
- Ejecutar gofmt sobre los tres archivos modificados.
- Ejecutar go test ./....
- Ejecutar go vet ./....
- Revisar el diff para confirmar que solo se añadieron o ajustaron comentarios.

## 7. Risks & Mitigations

- **Comentarios desactualizados respecto a la lógica actual:** describir exactamente el comportamiento existente, incluida la URL de PostgreSQL codificada, sin asumir que usa DATABASE_URL.
- **Exceso de comentarios:** agrupar la explicación por responsabilidad y limitar cada comentario a una o dos frases.
- **Cambios funcionales accidentales:** verificar el diff y ejecutar las pruebas después del formateo.
