## 1. Goal

Implementar estrictamente Phantom Token en el backend Go: exponer sólo tokens opacos al cliente, guardar el JWT firmado exclusivamente en Redis con TTL alineado a su expiración, validar firma y expiración antes de cada handler protegido, y añadir `POST /auth/login` y `POST /auth/refresh` con refresh tokens hasheados y rotación estricta.

## 2. Approach

Separar explícitamente tres responsabilidades que hoy están mezcladas: un emisor/verificador JWT interno, un almacén Redis de `token opaco -> JWT`, y un repositorio PostgreSQL de refresh tokens. El middleware recibirá únicamente el valor Bearer opaco, resolverá el JWT en Redis, lo verificará con el algoritmo y secreto configurados, y pondrá sólo la identidad validada en el contexto; los handlers de login/refresh devolverán DTOs con `access_token` opaco y `refresh_token`, nunca el JWT.

Se conservarán las capas hexagonales existentes: los puertos vivirán en domain/application, PostgreSQL y Redis seguirán como adaptadores, y `cmd/api/main.go` sólo compondrá dependencias y montará `/auth` como ruta pública. La rotación estricta usará un refresh token con selector público y secreto aleatorio: PostgreSQL buscará por selector, comparará el secreto con bcrypt, revocará el registro anterior dentro de una transacción y creará el nuevo registro.

## 3. File Changes

- **Modify** `internal/auth/domain/auth.go`: reemplazar el contrato actual que autentica claims directamente en Redis por puertos para emitir un JWT interno, crear/consultar/revocar sesiones Phantom y representar una identidad con `UserID`, `RoleID` y `ExpiresAt`; no exponer tipos JWT, Redis o pgx.
- **Create** `internal/auth/domain/refresh_token.go`: definir el contrato de refresh token y el modelo mínimo de persistencia (selector, usuario, hash, expiración, revocación) junto con errores sentinel para token inválido, expirado y reutilizado.
- **Modify** `internal/auth/application/authenticate_token.go`: hacer que el caso de uso reciba el opaco, obtenga el JWT desde el puerto Phantom, delegue la validación criptográfica al puerto JWT y consulte que el usuario siga activo; eliminar la aceptación de claims Redis como prueba suficiente.
- **Create** `internal/auth/application/login.go`: validar email/password contra `users/domain.UserRepository`, comparar bcrypt, rechazar usuario inexistente, contraseña inválida o inactivo sin filtrar cuál falló, emitir JWT interno, crear el opaco en Redis, crear el refresh hasheado en PostgreSQL y devolver sólo los dos tokens públicos.
- **Create** `internal/auth/application/refresh.go`: validar el refresh por selector/hash/expiración, cargar el usuario activo, emitir un JWT y opaco nuevos, rotar el refresh anterior dentro de la operación transaccional y devolver el nuevo par.
- **Create** `internal/auth/application/token_service.go` (o consolidar sus interfaces en los archivos anteriores): declarar los puertos consumidores para emisión/verificación JWT, emisión/lectura/revocación Phantom y refresh rotation, evitando que application dependa de implementaciones concretas.
- **Modify** `internal/auth/infrastructure/redis/token_store.go`: generar el valor opaco con `crypto/rand`, guardar exactamente el JWT firmado como valor de `phantom:<opaco>` mediante `SETEX/Set EX`, calcular TTL desde `exp - now`, resolver sólo el JWT durante autenticación y revocar eliminando la clave; no guardar claims como sustituto del JWT.
- **Create** `internal/auth/infrastructure/postgres/refresh_token_repository.go`: implementar consulta por selector, comparación bcrypt, inserción y rotación/revocación transaccional usando pgx; mapear `pgx.ErrNoRows` a error de refresh inválido y nunca devolver hashes al transporte.
- **Modify** `internal/users/infrastructure/jwt/issuer.go`: añadir verificación estricta de JWT interno, exigir HS256, validar firma y claims `user_id`, `role`, `iat`, `exp`, y exponer al application sólo una identidad tipada; mantener el JWT como dato interno.
- **Modify** `internal/users/domain/UserRepository.go`: conservar `GetByEmail` como puerto de login y documentar/ajustar el contrato para que el flujo pueda recuperar la contraseña hasheada y el estado activo sin exponer ese dato en DTOs.
- **Modify** `internal/auth/transport/http/handler.go`: añadir handlers `POST /login` y `POST /refresh`, validar JSON, mapear errores a `internal/platform/httpx`, y retirar cualquier respuesta que pueda transportar el JWT; mantener Google/logout/introspect sólo si siguen siendo parte del contrato, migrándolos al mismo emisor Phantom.
- **Modify** `internal/auth/transport/http/dto.go`: añadir request DTOs para email/password y refresh token, y response DTO `access_token` + `refresh_token`; no incluir JWT, claims sin necesidad, hashes ni valores Redis.
- **Modify** `internal/auth/transport/http/auth_middleware.go`: aceptar exactamente `Authorization: Bearer <token-opaco>`, resolver/verificar JWT antes de llamar a `next`, usar `errors.Is` para clasificar errores y no registrar ni devolver tokens.
- **Modify** `cmd/api/main.go`: construir `jwt.Issuer` con `JWT_SECRET` y expiración, conectar Redis Phantom, repositorio de refresh y casos de uso, montar el router público en `/auth`, y aplicar middleware a usuarios/mapas protegidos; no montar los endpoints de autenticación detrás del middleware.
- **Modify** `internal/platform/config/config.go`: validar `JWT_SECRET` con longitud segura, separar explícitamente expiración JWT de TTL si procede y conservar prefijo/TTL Redis; garantizar que el TTL efectivo nunca exceda `exp - now`.
- **Create** `migrations/004_refresh_tokens.sql`: crear tabla forward-only para refresh tokens con selector único, hash bcrypt, `user_id` FK, `expires_at`, `revoked_at`, `created_at` e índices necesarios; incluir sección down sólo si el tooling del proyecto la soporta.
- **Modify** `internal/users/application/create_user.go`, `internal/users/application/google_login.go` y `internal/users/transport/http/dto.go`: adaptar los flujos existentes para emitir el nuevo access token opaco mediante el servicio Phantom y evitar regresiones en los cambios locales ya presentes.
- **Modify** `internal/users/transport/http/handler.go`: mantener el endpoint de alta compatible con el contrato actual, pero verificar que el token devuelto sea opaco y que no se emita un JWT directo.
- **Create** `internal/auth/application/login_test.go` y `refresh_test.go`: pruebas table-driven de credenciales válidas/ inválidas, usuario inactivo, expiración, rotación, reuse del refresh anterior y ausencia de JWT en resultados.
- **Modify** `internal/auth/transport/http/handler_test.go` y **Create** `internal/auth/transport/http/auth_middleware_test.go`: probar status/body JSON, rutas `/auth/login` y `/auth/refresh`, Bearer ausente/malformado, JWT con firma incorrecta, JWT expirado y asegurar que ninguna respuesta contiene un JWT.
- **Create** `internal/auth/infrastructure/redis/token_store_test.go` y `internal/users/infrastructure/jwt/issuer_test.go`: comprobar clave/prefijo, valor JWT, TTL, revocación, algoritmo rechazado, firma inválida, expiración y claims inválidos.
- **Create** `internal/auth/infrastructure/postgres/refresh_token_repository_test.go` si el proyecto incorpora un harness PostgreSQL; de lo contrario, cubrir la transacción mediante integración en `test/` y dejar explícita la dependencia de PostgreSQL.
- **Modify** `migrations/README.md`: documentar la migración de refresh tokens y las variables `JWT_SECRET`, expiración, prefijo y TTL, sin incluir secretos.

## 4. Implementation Steps

### Task 1: Fijar el contrato y corregir el modelo actual

1. En `internal/auth/domain/auth.go` y `internal/auth/domain/refresh_token.go`, definir los puertos y errores para JWT, Phantom y refresh.
2. En `internal/users/infrastructure/jwt/issuer.go`, implementar emisión y verificación estricta; el parser debe rechazar algoritmos distintos de HS256, firma inválida, `exp` ausente/no numérico y expiración pasada.
3. En `internal/auth/infrastructure/redis/token_store.go`, cambiar el valor almacenado de claims serializados a JWT firmado y usar el TTL restante calculado desde la expiración del JWT.
4. En `internal/auth/application/authenticate_token.go`, encadenar opaco -> Redis -> JWT verificado -> usuario activo. El middleware no podrá continuar si cualquier etapa falla.

### Task 2: Persistir refresh tokens de forma segura

1. Crear `migrations/004_refresh_tokens.sql` con selector único, hash bcrypt, usuario, expiración y revocación.
2. Implementar `internal/auth/infrastructure/postgres/refresh_token_repository.go` con consultas parametrizadas y una transacción de rotación que invalide el token anterior antes de confirmar el nuevo.
3. Generar refresh tokens con suficiente entropía y formato selector/secreto; almacenar sólo selector y hash bcrypt, nunca el valor completo.
4. Mapear un refresh revocado, expirado, inexistente o reutilizado al mismo error público de credenciales inválidas, sin revelar cuál condición ocurrió.

### Task 3: Implementar login y refresh

1. En `internal/auth/application/login.go`, normalizar email, recuperar el usuario por `GetByEmail`, comparar con `bcrypt.CompareHashAndPassword`, exigir `EsActivo), crear JWT interno con claims de usuario/rol y emitir el opaco.
2. Persistir el refresh sólo después de tener una sesión de acceso válida; si falla Redis o PostgreSQL, devolver error interno y no devolver ningún token parcial.
3. En `internal/auth/application/refresh.go`, validar el refresh y el usuario activo, emitir un nuevo JWT/opaco y ejecutar rotación estricta del refresh anterior.
4. Reutilizar estas capacidades en Google/alta de usuario donde ya se emiten sesiones, evitando que esos flujos sigan devolviendo directamente otro formato de token.

### Task 4: Exponer HTTP y montar rutas

1. En `internal/auth/transport/http/dto.go`, definir JSON estable: request login `email/password`, request refresh `refresh_token`, response `access_token/refresh_token`.
2. En `internal/auth/transport/http/handler.go`, registrar `POST /login` y `POST /refresh`; responder 400 para body inválido, 401 para credenciales/token inválido, 403 para usuario inactivo y 500 para fallos de infraestructura.
3. En `internal/auth/transport/http/auth_middleware.go`, exigir Bearer opaco y colocar únicamente identidad validada en context.
4. En `cmd/api/main.go`, montar `authHandler.Routes()` bajo `/auth` en el router principal y mantener las rutas de negocio bajo middleware; revisar explícitamente que no se genere otra ruta pública accidental.
5. Actualizar Google/logout/introspect sólo para que consuman el mismo contrato Phantom, si se conservan en la superficie pública actual.

### Task 5: Pruebas y verificación de seguridad

1. Añadir pruebas unitarias table-driven de application, JWT, Redis y middleware.
2. Añadir pruebas HTTP que inspeccionen cada respuesta de login/refresh/google/alta y fallen si aparece una cadena con forma JWT (`xxxxx.yyyyy.zzzzz`) o un campo no permitido.
3. Ejecutar `gofmt` en cada Go modificado, `go test ./...` y `go vet ./...`.
4. Con PostgreSQL y Redis disponibles, probar manualmente login válido, credenciales inválidas, acceso protegido, refresh válido, reutilización del refresh anterior, expiración y revocación.

## 5. Acceptance Criteria

- `POST /auth/login` con email/password válidos responde 200 y JSON con exactamente `access_token` opaco y `refresh_token`; no contiene el JWT firmado.
- `POST /auth/login` compara la contraseña con bcrypt, rechaza contraseña incorrecta y no diferencia públicamente usuario inexistente de password incorrecto.
- Redis contiene `phantom:<access_token>` cuyo valor es el JWT interno firmado; el cliente nunca recibe ese valor.
- El TTL de la clave Redis es positivo y no mayor que `exp - now` del JWT; al expirar el JWT, la clave no permite autenticación.
- El middleware rechaza ausencia de Bearer, esquema incorrecto, opaco desconocido, Redis sin JWT, JWT expirado, firma incorrecta, algoritmo distinto de HS256 o claims inválidos antes de ejecutar el handler.
- Las rutas protegidas sólo reciben una identidad creada tras verificación criptográfica; no aceptan claims construidos únicamente desde Redis.
- `POST /auth/refresh` valida selector, hash bcrypt, expiración, revocación y usuario activo; ante éxito revoca el refresh anterior y devuelve un access token opaco y un refresh nuevo.
- Reutilizar el refresh anterior después de una rotación responde 401 y no crea otra sesión.
- PostgreSQL nunca almacena el refresh en claro y ninguna respuesta/log/error contiene su hash o el JWT.
- Un usuario inactivo no puede usar un access token o refresh token previamente emitido.
- Google y alta de usuario, si se mantienen, devuelven el mismo tipo de access token opaco y nunca un JWT.
- El contrato público de autenticación está bajo `/auth/login` y `/auth/refresh`; no se requiere el prefijo anterior `/api/v1/users/auth`.
- `go test ./...`, `go vet ./...` y `gofmt -l` sin archivos Go pendientes terminan correctamente.
- Domain/application no importan Redis, pgx, HTTP ni tipos JWT concretos; las dependencias apuntan hacia los puertos.

## 6. Verification Steps

1. Ejecutar `rg -n 'SignedString|Parse|JWT|token' internal/auth internal/users/transport cmd/api` y revisar que `SignedString` sólo se use en infraestructura y que ningún DTO de transporte contenga JWT.
2. Ejecutar `go test ./...` y `go vet ./...`; ejecutar `gofmt -l` y exigir salida vacía.
3. Levantar PostgreSQL y Redis, aplicar migraciones y arrancar con `JWT_SECRET` válido.
4. Enviar login válido y comprobar que `access_token` no tiene tres segmentos separados por puntos; inspeccionar Redis sólo desde administración para confirmar que el valor sí es el JWT.
5. Llamar una ruta protegida con el opaco: debe permitirla; repetir con JWT directo, opaco alterado, Bearer ausente y firma/expiración alteradas: todos deben responder 401.
6. Ejecutar refresh válido: comprobar access y refresh nuevos; repetir con el refresh anterior: 401; consultar PostgreSQL y confirmar hash, revocación y fechas.
7. Marcar el usuario inactivo y comprobar que tanto el opaco aún vigente como el refresh son rechazados.
8. Configurar TTL corto, esperar expiración y verificar que Redis y el middleware rechazan el acceso.
9. Confirmar que `/health` y `/health/ready` siguen disponibles y que ningún log/respuesta contiene el JWT interno.

## 7. Risks & Mitigations

- **El código actual no usa realmente el emisor JWT:** conectar `JWT_SECRET` en `cmd/api/main.go` y hacer fallar el arranque si falta o es débil; no tratar claims Redis como validación.
- **Cambios locales existentes en creación de usuarios:** revisar e integrar esos cambios sin sobrescribirlos; adaptar sólo los contratos necesarios y conservar sus pruebas.
- **Rotación concurrente del mismo refresh:** usar una transacción y actualización condicionada a `revoked_at IS NULL`; sólo una solicitud podrá ganar la rotación.
- **Desalineación JWT/Redis:** calcular `exp` una sola vez, derivar el TTL de ese mismo instante y rechazar TTL no positivo.
- **Filtración por introspect/logs o DTOs heredados:** auditar todos los handlers, eliminar campos de token JWT y añadir pruebas que inspeccionen cuerpos y logs.
- **Incompatibilidad con Flutter por cambio de ruta:** el contrato elegido es sólo `/auth/*`; documentar el cambio y actualizar el cliente antes del despliegue.
- **Fallo parcial entre Redis y PostgreSQL:** no devolver tokens hasta completar ambas persistencias; añadir revocación/compensación en errores de la segunda operación y registrar sólo IDs no sensibles.
