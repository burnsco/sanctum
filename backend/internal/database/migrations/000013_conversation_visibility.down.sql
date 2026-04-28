DROP INDEX IF EXISTS idx_conversations_visibility;

ALTER TABLE conversations
    DROP CONSTRAINT IF EXISTS chk_conversations_visibility;

ALTER TABLE conversations
    DROP COLUMN IF EXISTS visibility;
