# Migraciones

La migración `001_init.sql` crea:

- roles
- genero
- usuario

También habilita PostGIS.

La columna `fehcaact` conserva el nombre original
del diagrama proporcionado.

Si se desea corregir el nombre:

```sql
ALTER TABLE usuario
RENAME COLUMN fehcaact TO fechaact;