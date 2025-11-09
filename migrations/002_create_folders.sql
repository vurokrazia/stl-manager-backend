-- Migration: Create folders table
-- Description: Stores folders/directories that contain files

-- Up Migration
CREATE TABLE IF NOT EXISTS folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_folders_path ON folders(path);
CREATE INDEX idx_folders_name ON folders(name);

-- Down Migration
-- DROP TABLE IF EXISTS folders CASCADE;
