## 1. Objetivo

Mejorar la legibilidad y mantenibilidad del backend, añadiendo comentarios breves y corrigiendo inconsistencias funcionales de usuarios, paginación, configuración y esquema sin cambiar las rutas HTTP públicas.

## 2. Enfoque

Se normalizará primero el formato y la nomenclatura Go, y se extraerán operaciones repetidas de lectura de usuarios en el adaptador PostgreSQL para que los flujos sean más fáciles de seguir. La paginación se implementará en el repositorio mediante límites, desplazamientos y conteo, en vez de cargar todos los usuarios y devolver metadatos ficticios. El esquema inicial se alineará con los nombres usados por la aplicación y una migración de compatibilidad permitirá corregir instalaciones creadas con la versión actual.

Los comentarios serán concisos, en español y solo documentarán decisiones no obvias: precedencia de configuración, normalización/hash de credenciales, límites de paginación, semántica de disponibilidad y apagado ordenado. No se añadirán comentarios redundantes a getters, DTOs ni asignaciones evidentes.

## 3. Cambios de archivos

- **Modificar** `cmd/api/main.go` (líneas 26-260): reagrupar el cableado de dependencias y las rutas de salud; conservar el apagado ordenado y añadir comentarios cortos para disponibilidad y cierre del servidor.
- **Modificar** `internal/platform/config/config.go` (líneas 12-80): retirar mensajes de consola y comentarios contradictorios; cargar `.env` sin sobrescribir el entorno, ignorar únicamente su ausencia y propagar un error contextual cuando no pueda leerse por otra causa.
- **Modificar** `internal/platform/database/postgres.go` (líneas 10-48): documentar brevemente la política de renovación/validación del pool y envolver fallos de parseo, creación y ping con contexto.
- **Modificar** `internal/platform/redis/client.go` (líneas 9-24): envolver el error al interpretar la URL de Redis y documentar que `Ping` se reutiliza para la prueba de disponibilidad.
- **Modificar** `internal/platform/httpx/json.go` (líneas 8-39): ordenar el formato y dejar explícita la respuesta JSON de error; conservar el contrato `{"error": ...}`.
- **Modificar** `internal/users/domain/UserRepository.go` (líneas 5-12): renombrar `GetById` a `GetByID`; sustituir `List(ctx)` por operaciones paginadas y de conteo, manteniendo el puerto definido en el dominio.
- **Modificar** `internal/users/domain/user.go`, `role.go`, `genero.go` y `errors.go` (líneas 3-20): aplicar formato idiomático y comentarios de paquete/tipo solo cuando aclaren el papel del modelo o de los errores centinela; no cambiar los campos expuestos ni los valores de error.
- **Modificar** `internal/users/application/create_user.go` (líneas 11-72): centralizar la normalización de texto, validar IDs de rol y género positivos antes de consultar/persistir, documentar el hash de contraseña y envolver los fallos de infraestructura preservando `errors.Is`.
- **Modificar** `internal/users/application/get_user.go`, `update_user.go` y `deactive_user.go` (líneas 9-59, 9-58 y 9-33): usar `GetByID`, validar identificadores antes de acceder al repositorio, mantener la normalización de nombres y añadir contexto a errores delegados.
- **Modificar** `internal/users/application/list_user.go` (líneas 9-67): eliminar líneas vacías y el offset comentado; calcular `offset`, consultar el total y la página a través del puerto, y calcular `TotalPages` por redondeo hacia arriba.
- **Modificar** `internal/users/infrastructure/postgres/user_repository.go` (líneas 12-365): implementar el puerto actualizado, usar el nombre de columna `fechaact`, extraer el escaneo común de usuario/rol/género donde sea aplicable, añadir `LIMIT/OFFSET` al listado y una consulta `COUNT(*)`; envolver errores de pgx y mantener `ErrUserNotFound` cuando no haya filas afectadas.
- **Modificar** `internal/users/transport/http/dto.go` (líneas 9-117): compactar las declaraciones de DTO y conservar el mapeo explícito que impide exponer la contraseña; añadir una nota corta sobre ese límite de transporte si resulta necesaria.
- **Modificar** `internal/users/transport/http/handler.go` (líneas 15-389): mejorar el formato, usar un decodificador que acepte exactamente un objeto JSON y rechace contenido extra, conservar el mapeo actual de errores a códigos HTTP y documentar brevemente las funciones auxiliares solo donde no sea obvio.
- **Modificar** `migrations/001_init.sql` (tabla `usuario`, líneas 12-32): declarar la tabla `roles` que el repositorio consulta, corregir los nombres de apellidos y `fechaact`, y declarar la relación de `usuario.idrol` con `roles.id` para que una base nueva coincida con la aplicación.
- **Crear** `migrations/002_align_user_schema.sql`: migración idempotente y forward-only para instalaciones existentes; creará `roles` si falta, renombrará los nombres de columna históricos cuando existan y añadirá las columnas de apellido que falten, sin borrar datos.
- **Modificar** `migrations/README.md` (líneas 1-18): describir con precisión las dos migraciones y eliminar la instrucción que propone un nombre de columna distinto al usado por el código.
- **Modificar** `docker-compose.yml` (líneas 11-18): montar el directorio de migraciones en vez de un único archivo para que una base nueva ejecute también la migración de alineación; ajustar el healthcheck para usar las mismas credenciales y base definidas en el servicio.
- **Crear** `internal/users/application/create_user_test.go`: pruebas table-driven con repositorio falso para normalización, contraseña corta, IDs inválidos, correo duplicado, hash de contraseña y propagación de errores.
- **Crear** `internal/users/application/list_user_test.go`: pruebas table-driven de valores por defecto, máximo de 100, offset enviado al repositorio, total y redondeo de páginas.
- **Crear** `internal/users/transport/http/handler_test.go`: pruebas HTTP con dobles de casos de uso para JSON desconocido o concatenado, validación de ID y mapeo de errores de dominio, sin requerir PostgreSQL ni Redis.

## 4. Pasos de implementación

### Tarea 1: Alinear el modelo y el puerto de usuarios

1. En `internal/users/domain/UserRepository.go`, cambiar el método de lectura a `GetByID` y definir `List(ctx, limit, offset)` junto con `Count(ctx)`.
2. En los casos de uso de `internal/users/application/`, actualizar las llamadas al puerto; eliminar el código comentado de `list_user.go` y calcular primero el total, después la página y por último `TotalPages = ceil(total/pageSize)`.
3. En `create_user.go` y `update_user.go`, aplicar la misma validación de nombre, apellido paterno, rol y género antes de persistir; normalizar correo únicamente al crear y hashear una contraseña solo después de validar.
4. Añadir pruebas de casos de uso con un fake de `UserRepository`, verificando argumentos de límite/offset y que las contraseñas nunca llegan al repositorio en texto plano.

### Tarea 2: Corregir y simplificar el adaptador PostgreSQL

1. En `internal/users/infrastructure/postgres/user_repository.go`, adaptar las firmas al nuevo puerto y cambiar todas las referencias `fehcaact` por `fechaact`.
2. Extraer funciones privadas para construir/escanear el modelo de usuario con las relaciones que se devuelven en detalle y listado, sin filtrar tipos de pgx fuera del adaptador.
3. Implementar el listado ordenado con `LIMIT $1 OFFSET $2` y el conteo total con `COUNT(*)`; conservar la detección de `pgx.ErrNoRows` y de cero filas afectadas.
4. Añadir contexto con `fmt.Errorf(... %w)` a cada error de operación de base de datos para que el origen sea diagnosticable sin romper `errors.Is`.

### Tarea 3: Fortalecer los adaptadores de plataforma y HTTP

1. En `internal/platform/config/config.go`, preservar la precedencia de variables de entorno, silenciar el caso normal de un `.env` ausente y devolver errores contextuales para otros fallos de carga y configuración inválida.
2. En `internal/platform/database/postgres.go` y `internal/platform/redis/client.go`, contextualizar los errores y añadir solo los comentarios sobre políticas de conexión o disponibilidad.
3. En `internal/users/transport/http/handler.go`, reforzar `decodeJSON` para rechazar un segundo valor JSON; mantener cuerpos de error y status codes existentes.
4. En `cmd/api/main.go`, agrupar la composición de dependencias y las rutas; documentar el propósito de `/health`, `/health/ready` y el timeout de apagado sin trasladar reglas de negocio al transporte.
5. En `dto.go` y `httpx/json.go`, aplicar formato compacto y preservar el contrato de respuesta, en particular la exclusión de `Password`.

### Tarea 4: Alinear la inicialización de datos

1. Corregir `migrations/001_init.sql` para que las instalaciones nuevas creen `roles`, `usuario.apepat`, `usuario.apemat` y `usuario.fechaact` de acuerdo con las consultas.
2. Crear `migrations/002_align_user_schema.sql` con comprobaciones condicionales para transformar de manera segura las columnas históricas, sin editar ni eliminar registros existentes.
3. Actualizar el montaje de Docker y el healthcheck en `docker-compose.yml`, y actualizar `migrations/README.md` para que la documentación no recomiende un nombre incompatible.

### Tarea 5: Formato y verificación

1. Ejecutar `gofmt` sobre cada archivo Go modificado.
2. Ejecutar `go test ./...` y `go vet ./...`.
3. Para una base nueva, iniciar los servicios y comprobar que ambos endpoints de salud responden correctamente; comprobar que `GET /api/v1/users?page=2&pageSize=20` devuelve la segunda página y metadatos reales.
4. Verificar manualmente que crear usuarios normaliza correo, guarda un hash bcrypt, y que un JSON con campos desconocidos o dos objetos devuelve HTTP 400.

## 5. Criterios de aceptación

- Todos los archivos Go modificados están formateados por `gofmt`, y `go vet ./...` termina sin diagnósticos.
- `go test ./...` termina correctamente e incluye cobertura de los flujos de creación y paginación de usuarios.
- `GET /api/v1/users?page=P&pageSize=S` consulta como máximo S registros, usa offset `(P-1)*S`, devuelve `Total` igual al conteo de usuarios y `TotalPages = ceil(Total/S)`.
- Un `page` menor o igual a cero se convierte en 1; un `pageSize` menor o igual a cero se convierte en 20; uno mayor que 100 se limita a 100.
- Crear o actualizar un usuario con rol o género no positivo responde HTTP 400; crear un usuario válido persiste una contraseña bcrypt, no la contraseña original.
- Las rutas existentes y sus métodos HTTP se conservan: `POST/GET /api/v1/users/`, `GET/PUT/DELETE /api/v1/users/{id}`, `/health` y `/health/ready`.
- Un cuerpo con campo desconocido, JSON malformado o contenido JSON adicional recibe HTTP 400.
- Una base inicializada desde cero contiene `roles` y las columnas de `usuario` que usa el repositorio; el arranque no falla por `apepat`, `apemat`, `fechaact` o ausencia de `roles`.
- No quedan comentarios de código desactivado; los comentarios nuevos explican una decisión concreta en una o dos líneas.

## 6. Pasos de verificación

1. Ejecutar `gofmt -w` solo sobre los archivos Go cambiados y revisar `git diff --check` para detectar espacios o marcadores accidentales.
2. Ejecutar `go test ./...` y `go vet ./...`.
3. Con una instancia limpia de PostgreSQL/Redis, ejecutar `docker compose up -d` y `go run ./cmd/api`; consultar `/health` y `/health/ready` y esperar HTTP 200.
4. Crear más de 20 usuarios de prueba y comprobar las páginas 1, 2, tamaño inválido, tamaño 101 y página 0; validar tanto `items` como `total` y `totalPages`.
5. Enviar una creación con contraseña de menos de ocho caracteres, rol/género cero, correo duplicado, campo JSON extra y dos objetos concatenados; comprobar HTTP 400 o 409 según corresponda.
6. Aplicar la migración de compatibilidad a una copia de la estructura anterior y comprobar que las consultas de crear, obtener, listar, actualizar y desactivar usuarios ya no fallan por nombres de tabla o columna.

## 7. Riesgos y mitigaciones

- **Migración ya aplicada en entornos existentes:** editar solo la inicial corrige bases nuevas, por eso se añade `002_align_user_schema.sql` idempotente que transforma estructura sin borrar datos.
- **Datos previos con un único apellido en `apermat`:** la migración lo conserva al renombrarlo a `apepat` y crea `apemat` vacío para no inventar datos.
- **Error de paginación por cambios del puerto:** las pruebas con fake verifican límite, offset, conteo y cálculo de páginas antes de integrar PostgreSQL.
- **Cambios en errores envueltos:** se usará `%w` y `errors.Is`, por lo que el handler mantiene el mapeo HTTP actual.
- **Arranque local con volumen Docker existente:** Docker no reejecuta scripts de inicialización en un volumen ya creado; la verificación especificará aplicar la migración de compatibilidad al entorno existente o usar una base limpia.