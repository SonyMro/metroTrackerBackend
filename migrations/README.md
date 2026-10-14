# Migraciones

`001_init.sql` crea el esquema para instalaciones nuevas, incluidas las tablas `roles`, `genero` y `usuario` con las columnas `apepat`, `apemat` y `fechaact`.

`002_align_user_schema.sql` actualiza de forma idempotente instalaciones previas: conserva datos al renombrar columnas históricas y agrega las columnas o relación faltantes.

`003_google_auth.sql` agrega `usuario.google_id` y la tabla `user_auth`, con unicidad para `(provider, provider_id)` y para el vínculo de un usuario con cada proveedor. El endpoint `POST /api/v1/users/auth/google` requiere `GOOGLE_CLIENT_ID`, `JWT_SECRET`, `JWT_EXPIRATION` (por ejemplo `24h`), `DEFAULT_ROLE_ID` y `DEFAULT_GENDER_ID`.

`004_refresh_tokens.sql` crea el almacenamiento de refresh tokens por selector y hash bcrypt. Configura `JWT_SECRET` (mínimo 32 caracteres), `JWT_EXPIRATION`, `AUTH_TOKEN_TTL` y `AUTH_TOKEN_PREFIX`; Redis siempre usa como TTL efectivo el menor tiempo restante del JWT.
