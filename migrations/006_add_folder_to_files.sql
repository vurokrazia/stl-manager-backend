-- Migration: Add folder_id to files table
-- Description: Links files to their parent folder (nullable for root-level files)

-- Up Migration
ALTER TABLE files ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;

CREATE INDEX idx_files_folder ON files(folder_id);

-- Down Migration
-- ALTER TABLE files DROP COLUMN IF EXISTS folder_id;
