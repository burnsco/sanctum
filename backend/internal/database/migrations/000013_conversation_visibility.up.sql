ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) NOT NULL DEFAULT 'direct';

UPDATE conversations
SET visibility = 'public'
WHERE is_group = TRUE AND sanctum_id IS NOT NULL;

UPDATE conversations
SET visibility = 'private'
WHERE is_group = TRUE AND sanctum_id IS NULL;

UPDATE conversations
SET visibility = 'direct'
WHERE is_group = FALSE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_conversations_visibility'
    ) THEN
        ALTER TABLE conversations
            ADD CONSTRAINT chk_conversations_visibility
            CHECK (visibility IN ('public', 'private', 'direct'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_conversations_visibility ON conversations (visibility);
