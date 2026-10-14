-- Alinea instalaciones creadas con versiones anteriores sin borrar datos de usuario.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'apermat')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'apepat') THEN
        ALTER TABLE usuario RENAME COLUMN apermat TO apepat;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'fehcaact')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'fechaact') THEN
        ALTER TABLE usuario RENAME COLUMN fehcaact TO fechaact;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'apepat') THEN
        ALTER TABLE usuario ADD COLUMN apepat VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'apemat') THEN
        ALTER TABLE usuario ADD COLUMN apemat VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'usuario' AND column_name = 'fechaact') THEN
        ALTER TABLE usuario ADD COLUMN fechaact TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

ALTER TABLE usuario DROP CONSTRAINT IF EXISTS usuario_idrol_fkey;
ALTER TABLE usuario DROP COLUMN IF EXISTS idrol;
DROP TABLE IF EXISTS roles;
