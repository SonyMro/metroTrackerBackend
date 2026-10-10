## 1. Goal

Agregar al módulo de users un flujo de login con Google que valide el `idToken`, resuelva o cree/vincule usuarios de forma transaccional, y devuelva un JWT propio junto con la información básica del usuario.

## 2. Approach

Se añadirá un caso de uso específico para Google Login, manteniendo la validación del token y la firma JWT detrás de puertos/interfaces para que la aplicación no dependa directamente de HTTP ni de PostgreSQL. El adaptador PostgreSQL resolverá la búsqueda por proveedor, la vinculación por correo y la creación de `usuario` + `user_auth` dentro de una única transacción; el handler sólo decodificará el DTO, invocará el caso de uso y mapeará la respuesta. Se incluirá una migración idempotente/forward-only para `google_id` y `user_auth`, con constraints e índices que protejan duplicados.

## 3. File Changes

- **Modify** [go.mod](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/go.mod?type=file&root=%252F): agregar `google.golang.org/api/idtoken` y `github.com/golang-jwt/jwt/v5`.
- **Modify** [internal/platform/config/config.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/config/config.go?type=file&root=%252F): incorporar `GOOGLE_CLIENT_ID`, `JWT_SECRET`, expiración del token y `DEFAULT_ROLE_ID`/`DEFAULT_GENDER_ID`; validar que los secretos, el client ID y los IDs sean utilizables antes de arrancar.
- **Modify** [internal/users/domain/user.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/domain/user.go?type=file&root=%252F): añadir el identificador de Google si debe formar parte del modelo persistido y mantener el DTO de usuario separado del token.
- **Modify** [internal/users/domain/UserRepository.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/domain/UserRepository.go?type=file&root=%252F): definir los puertos necesarios para buscar el vínculo `provider/provider_id`, buscar por correo y ejecutar la resolución/creación de identidad Google de manera transaccional.
- **Modify** [internal/users/domain/errors.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/domain/errors.go?type=file&root=%252F): agregar errores distinguibles para token inválido, configuración inválida, vínculo duplicado y cuenta inactiva, para que el transporte produzca respuestas HTTP correctas.
- **Create** [internal/users/application/google_login.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/application/google_login.go?type=file&root=%252F): implementar el caso de uso; validar el token mediante un puerto, extraer email/nombre/subject, normalizar el correo, resolver el vínculo/correo/alta, y solicitar la emisión del JWT propio.
- **Create** [internal/users/application/google_login_test.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/application/google_login_test.go?type=file&root=%252F): pruebas table-driven para vínculo existente, vinculación por correo, creación, token inválido, datos incompletos y propagación de errores.
- **Create** [internal/users/infrastructure/google/idtoken_verifier.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/google/idtoken_verifier.go?type=file&root=%252F): adaptar `google.golang.org/api/idtoken.Validate` con `GOOGLE_CLIENT_ID`, exponiendo sólo los claims requeridos al caso de uso.
- **Create** [internal/users/infrastructure/jwt/issuer.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/jwt/issuer.go?type=file&root=%252F): firmar el JWT de la aplicación con `golang-jwt/jwt/v5`, incluyendo `user_id`, `role`, `iat`, `exp` y un algoritmo HMAC explícito.
- **Modify** [internal/users/infrastructure/postgres/user_repository.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/postgres/user_repository.go?type=file&root=%252F): implementar la consulta de `user_auth`, la búsqueda por correo y una operación transaccional que inserte usuario y vínculo, haga rollback ante cualquier error y recupere el usuario con su rol.
- **Modify** [internal/users/transport/http/dto.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/transport/http/dto.go?type=file&root=%252F): agregar request `idToken` y response de login con `token` e información básica del usuario.
- **Modify** [internal/users/transport/http/handler.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/transport/http/handler.go?type=file&root=%252F): añadir `POST /auth/google`, validar JSON, invocar el caso de uso y mapear errores de token/credenciales sin exponer detalles internos.
- **Modify** [cmd/api/main.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/cmd/api/main.go?type=file&root=%252F): construir los verificadores/emisor JWT con configuración y conectarlos al handler de users.
- **Create** [migrations/003_google_auth.sql](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/migrations/003_google_auth.sql?type=file&root=%252F): agregar `usuario.google_id` y crear `user_auth` con FK a `usuario`, `provider`, `provider_id`, índices/constraints de unicidad y valores compatibles con el esquema existente.
- **Modify** [migrations/README.md](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/migrations/README.md?type=file&root=%252F): documentar la nueva migración, sus columnas y variables de configuración requeridas.
- **Modify** [internal/users/transport/http/handler_test.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/transport/http/handler_test.go?type=file&root=%252F): cubrir body inválido, ausencia de `idToken` y respuesta exitosa/error del nuevo endpoint, ajustando la construcción del handler si se abstrae el caso de uso.

## 4. Implementation Steps

### Task 1: Configuration and dependency ports

1. En [internal/platform/config/config.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/platform/config/config.go?type=file&root=%252F), leer `GOOGLE_CLIENT_ID`, `JWT_SECRET`, `JWT_EXPIRATION`, `DEFAULT_ROLE_ID` y `DEFAULT_GENDER_ID`; rechazar configuración ausente o inválida.
2. En [internal/users/domain/UserRepository.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/domain/UserRepository.go?type=file&root=%252F), declarar contratos para claims Google, verificación de identidad y emisión de token sin importar PostgreSQL, Google ni JWT.
3. En [go.mod](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/go.mod?type=file&root=%252F), añadir las dos dependencias y actualizar [go.sum](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/go.sum?type=file&root=%252F) mediante el gestor de módulos.

### Task 2: Google validation and JWT issuance

1. En [internal/users/infrastructure/google/idtoken_verifier.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/google/idtoken_verifier.go?type=file&root=%252F), llamar a `idtoken.Validate(ctx, rawToken, clientID)`, rechazar claims sin `email` o `sub`, y mapear `email`, `name` y `sub`.
2. En [internal/users/infrastructure/jwt/issuer.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/jwt/issuer.go?type=file&root=%252F), crear claims con `user_id` y `role`, establecer expiración configurable, firmar con HS256 y devolver errores contextualizados.
3. Añadir pruebas unitarias de mapeo/validación del caso de uso usando dobles de los puertos; la llamada real a Google se mantendrá en el adaptador y no se dependerá de red en las pruebas.

### Task 3: Persistence and transaction

1. En [migrations/003_google_auth.sql](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/migrations/003_google_auth.sql?type=file&root=%252F), crear `user_auth` y `usuario.google_id`; garantizar unicidad por `(provider, provider_id)` y, si corresponde al modelo elegido, unicidad de `usuario.google_id`.
2. En [internal/users/infrastructure/postgres/user_repository.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/infrastructure/postgres/user_repository.go?type=file&root=%252F), usar `BeginTx`, consultar primero el vínculo Google, después el correo, y bloquear/coordinar las lecturas necesarias para evitar carreras de vinculación.
3. Para cuenta existente, insertar sólo el vínculo con `provider = 'google'`; para cuenta nueva, insertar `usuario` con nombre, correo, `google_id`, `esactivo = TRUE`, IDs por defecto y luego insertar `user_auth`.
4. Hacer rollback ante error en cualquiera de las inserciones y commit sólo después de crear el vínculo; devolver el usuario completo con su rol para construir la respuesta.
5. Mantener el flujo de usuario/password existente sin alterar sus contratos.

### Task 4: Application and HTTP integration

1. En [internal/users/application/google_login.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/application/google_login.go?type=file&root=%252F), orquestar verificación, normalización y persistencia; no aceptar un usuario inactivo y no devolver el token de Google al cliente.
2. En [internal/users/transport/http/dto.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/transport/http/dto.go?type=file&root=%252F), definir `GoogleLoginRequest{IDToken}` y `GoogleLoginResponse{Token, User}`.
3. En [internal/users/transport/http/handler.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/internal/users/transport/http/handler.go?type=file&root=%252F), registrar `POST /auth/google`, exigir `idToken` no vacío y responder `200` con el JWT y usuario básico.
4. En [cmd/api/main.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/cmd/api/main.go?type=file&root=%252F), inyectar configuración, verificador, emisor y caso de uso en el handler.

### Task 5: Documentation and tests

1. Documentar en comentarios Go el propósito de cada puerto, el orden de resolución de identidad y la garantía transaccional.
2. Actualizar [migrations/README.md](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/migrations/README.md?type=file&root=%252F) con las variables de entorno y el endpoint.
3. Añadir pruebas table-driven para las cuatro ramas: vínculo existente, correo existente, usuario nuevo y fallos con rollback/error.
4. Ejecutar `gofmt` en cada archivo Go modificado y la verificación indicada abajo.

## 5. Acceptance Criteria

- Un `POST /api/v1/users/auth/google` con `{"idToken":"..." }` válido se valida mediante `google.golang.org/api/idtoken` contra `GOOGLE_CLIENT_ID`.
- Un token válido sin `email` o `sub` no crea ni vincula usuarios y devuelve un error 4xx.
- Si existe `user_auth(provider='google', provider_id=sub)`, se reutiliza exactamente su `user_id` y no se crea otra cuenta.
- Si no existe el vínculo pero sí `usuario.correo=email`, se crea un único vínculo Google para el usuario existente.
- Si no existe ni vínculo ni correo, se crea un usuario activo con nombre, correo, `google_id`, IDs por defecto y su registro `user_auth`.
- Si falla la inserción de `user_auth`, la transacción hace rollback y no queda un usuario nuevo huérfano.
- La respuesta exitosa contiene un JWT propio y la información básica del usuario, incluyendo `id`, nombre, correo y rol; no contiene el idToken original ni la contraseña.
- El JWT contiene `user_id`, `role`, `iat` y `exp`, está firmado con el secreto configurado y rechaza configuración sin secreto.
- Los constraints de base de datos impiden duplicar el mismo proveedor/provider_id y el mismo vínculo usuario-proveedor.
- Los flujos existentes de creación, consulta, actualización, listado y desactivación de users conservan su comportamiento actual.
- El código Go modificado queda formateado y los tests/vet del proyecto no introducen fallos atribuibles al cambio.

## 6. Verification Steps

1. Ejecutar `gofmt -w` sobre los archivos Go modificados.
2. Ejecutar `go test ./internal/users/...` para las pruebas del módulo.
3. Ejecutar `go test ./...` y `go vet ./...`.
4. Aplicar la migración en una base PostgreSQL de prueba y verificar columnas, FK, índices y constraints de `user_auth`.
5. Probar manualmente con un idToken válido emitido para `GOOGLE_CLIENT_ID`: primer login, segundo login, vinculación de correo existente y token con subject distinto.
6. Simular un fallo de inserción de `user_auth` en una prueba de integración y comprobar que el conteo de `usuario` no aumenta.
7. Decodificar el JWT de respuesta sin confiar en sus claims y comprobar `user_id`, `role`, `iat`, `exp`; verificar que la firma sólo valida con `JWT_SECRET`.
8. Verificar respuestas 400 para JSON inválido/idToken vacío, 401/403 para token Google inválido o cuenta inactiva, y 500 para fallas inesperadas de infraestructura.

## 7. Risks & Mitigations

- **Esquema existente divergente:** las migraciones actuales no contienen `user_auth` ni `google_id`; la migración nueva debe usar comprobaciones idempotentes y documentar supuestos de tipos/nombres.
- **Carreras de primer login:** dos solicitudes simultáneas pueden intentar vincular el mismo correo; constraints únicos más una transacción con aislamiento/bloqueo y tratamiento de conflicto evitan duplicados.
- **Nombres obligatorios del esquema:** `usuario.apepat` es NOT NULL, pero Google sólo garantiza un nombre libre; el caso de uso debe aplicar una regla explícita y documentada para completar campos heredados sin inventar datos sensibles.
- **Cuenta Google ya vinculada a otro correo:** el índice único por proveedor/provider_id debe producir un conflicto controlado, nunca reasignar silenciosamente el vínculo.
- **Secreto JWT mal configurado:** validar longitud/presencia al arranque y no registrar el secreto ni el idToken en logs.
- **Cambios en el constructor del handler:** actualizar los tests existentes y mantener la composición en [cmd/api/main.go](air-file://qapd8iufe89em8f96n4h/Users/marioguzmanperez/Documents/MetroTracker/MetroTrackerBackend/cmd/api/main.go?type=file&root=%252F) para evitar dependencias globales.

