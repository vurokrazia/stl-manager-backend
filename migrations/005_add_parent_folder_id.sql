-- Migration: Add parent_folder_id to support hierarchical folders
-- Description: Allows folders to contain other folders recursively

-- Up Migration
ALTER TABLE folders
ADD COLUMN parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE;

CREATE INDEX idx_folders_parent_folder_id ON folders(parent_folder_id);

-- Down Migration
-- ALTER TABLE folders DROP COLUMN parent_folder_id;
-- DROP INDEX IF EXISTS idx_folders_parent_folder_id;
