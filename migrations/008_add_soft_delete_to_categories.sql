-- Add soft delete column to categories
ALTER TABLE categories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Add index for performance when filtering out deleted records
CREATE INDEX IF NOT EXISTS categories_deleted_at_idx ON categories(deleted_at) WHERE deleted_at IS NULL;
