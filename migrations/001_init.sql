-- Enable extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS citext;

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name CITEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Folders table
CREATE TABLE IF NOT EXISTS folders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  path TEXT UNIQUE NOT NULL,
  file_name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('stl','zip','rar')),
  size BIGINT NOT NULL,
  modified_at TIMESTAMPTZ NOT NULL,
  sha256 TEXT,
  folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Files-Categories junction table (N-N)
CREATE TABLE IF NOT EXISTS files_categories (
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (file_id, category_id)
);

-- Folders-Categories junction table (N-N)
CREATE TABLE IF NOT EXISTS folders_categories (
  folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (folder_id, category_id)
);

-- Scans table
CREATE TABLE IF NOT EXISTS scans (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status TEXT NOT NULL,
  found INT DEFAULT 0,
  processed INT DEFAULT 0,
  progress INT DEFAULT 0,
  error TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS files_name_trgm_idx ON files USING GIN (file_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS files_path_trgm_idx ON files USING GIN (path gin_trgm_ops);
CREATE INDEX IF NOT EXISTS files_type_idx ON files(type);
CREATE INDEX IF NOT EXISTS files_folder_idx ON files(folder_id);
CREATE INDEX IF NOT EXISTS folders_path_idx ON folders(path);
CREATE INDEX IF NOT EXISTS folders_name_idx ON folders(name);
CREATE INDEX IF NOT EXISTS folders_parent_folder_id_idx ON folders(parent_folder_id);
CREATE INDEX IF NOT EXISTS folders_categories_folder_idx ON folders_categories(folder_id);
CREATE INDEX IF NOT EXISTS folders_categories_category_idx ON folders_categories(category_id);
CREATE INDEX IF NOT EXISTS scans_status_idx ON scans(status);

-- Seed initial categories
INSERT INTO categories(name) VALUES
 ('figurine'),('miniature'),('mechanical_part'),('spare_part'),('tool_holder'),
 ('printer_upgrade'),('rc_part'),('cosplay_prop'),('terrain_piece'),('calibration'),
 ('enclosure_part'),('storage'),('stand'),('holder'),('clip'),
 ('adapter'),('mount'),('logo'),('keychain'),('phone_accessory'),
 ('vehicle'),('anime'),('character'),('diorama'),('furniture'),
 ('uncategorized')
ON CONFLICT (name) DO NOTHING;
