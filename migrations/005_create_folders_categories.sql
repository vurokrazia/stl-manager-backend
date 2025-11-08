-- Migration: Create folders_categories join table
-- Description: Many-to-many relationship between folders and categories

-- Up Migration
CREATE TABLE IF NOT EXISTS folders_categories (
    folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (folder_id, category_id)
);

CREATE INDEX idx_folders_categories_folder ON folders_categories(folder_id);
CREATE INDEX idx_folders_categories_category ON folders_categories(category_id);

-- Down Migration
-- DROP TABLE IF EXISTS folders_categories CASCADE;
