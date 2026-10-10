-- Identidad Google. Todas las operaciones son idempotentes para instalaciones existentes.
ALTER TABLE usuario ADD COLUMN IF NOT EXISTS google_id VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS usuario_google_id_uq ON usuario (google_id) WHERE google_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_auth (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_auth_provider_id_uq UNIQUE (provider, provider_id),
    CONSTRAINT user_auth_user_provider_uq UNIQUE (user_id, provider)
);

CREATE INDEX IF NOT EXISTS user_auth_user_id_idx ON user_auth (user_id);
