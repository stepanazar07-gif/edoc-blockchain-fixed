CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE users ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS age INTEGER;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

DO $$
DECLARE
    has_username BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'username'
    ) INTO has_username;

    IF has_username THEN
        EXECUTE 'ALTER TABLE users ALTER COLUMN username DROP NOT NULL';
        EXECUTE 'UPDATE users SET name = COALESCE(NULLIF(name, ''''), NULLIF(username, ''''), ''User'') WHERE name IS NULL OR name = ''''';
        EXECUTE 'UPDATE users SET phone = LEFT(COALESCE(NULLIF(phone, ''''), NULLIF(username, ''''), id::text), 32) WHERE phone IS NULL OR phone = ''''';
    ELSE
        UPDATE users SET name = COALESCE(NULLIF(name, ''), 'User') WHERE name IS NULL OR name = '';
        UPDATE users SET phone = LEFT(COALESCE(NULLIF(phone, ''), id::text), 32) WHERE phone IS NULL OR phone = '';
    END IF;

    UPDATE users SET age = 18 WHERE age IS NULL OR age < 1 OR age > 120;
END $$;

ALTER TABLE users ALTER COLUMN name SET NOT NULL;
ALTER TABLE users ALTER COLUMN age SET NOT NULL;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone);

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_hash TEXT NOT NULL,
    file_path TEXT NOT NULL,
    upload_date DATE NOT NULL DEFAULT CURRENT_DATE,
    upload_time TIME NOT NULL DEFAULT CURRENT_TIME,
    file_size BIGINT NOT NULL CHECK (file_size >= 0),
    mime_type TEXT NOT NULL DEFAULT 'application/octet-stream'
);

CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files(owner_id);
CREATE INDEX IF NOT EXISTS idx_files_file_hash ON files(file_hash);

CREATE TABLE IF NOT EXISTS file_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    transfer_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE,
    CHECK (sender_id <> receiver_id)
);

CREATE INDEX IF NOT EXISTS idx_file_transfers_receiver_id ON file_transfers(receiver_id);
CREATE INDEX IF NOT EXISTS idx_file_transfers_sender_id ON file_transfers(sender_id);
CREATE INDEX IF NOT EXISTS idx_file_transfers_status ON file_transfers(status);

DO $$
BEGIN
    IF to_regclass('public.documents') IS NOT NULL THEN
        INSERT INTO files (id, owner_id, file_name, file_hash, file_path, upload_date, upload_time, file_size, mime_type)
        SELECT id, owner_id, file_name, file_hash, './uploads/' || file_hash,
               uploaded_at::date, uploaded_at::time,
               file_size, COALESCE(mime_type, 'application/octet-stream')
        FROM documents
        ON CONFLICT (id) DO NOTHING;
    END IF;

    IF to_regclass('public.transfers') IS NOT NULL THEN
        INSERT INTO file_transfers (id, file_id, sender_id, receiver_id, status, transfer_date, accepted_at)
        SELECT id, document_id, from_user_id, to_user_id,
               CASE
                   WHEN status IN ('claimed', 'accepted') THEN 'accepted'
                   WHEN status = 'declined' THEN 'declined'
                   ELSE 'pending'
               END,
               created_at,
               claimed_at
        FROM transfers
        ON CONFLICT (id) DO NOTHING;
    END IF;
END $$;
