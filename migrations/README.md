# Migraciones

`001_init.sql` crea el esquema para instalaciones nuevas, incluidas las tablas `roles`, `genero` y `usuario` con las columnas `apepat`, `apemat` y `fechaact`.

`002_align_user_schema.sql` actualiza de forma idempotente instalaciones previas: conserva datos al renombrar columnas históricas y agrega las columnas o relación faltantes.
