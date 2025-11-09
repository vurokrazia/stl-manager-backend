# Database Migrations

**All migrations are in this directory (`migrations/`) in sequential order.**

## Migration Order

Migrations are numbered sequentially and applied in order:

1. **`001_init.sql`** - Creates base tables
   - Creates: `scans`, `files`, `categories`, `files_categories`
   - Enables: uuid-ossp, pg_trgm extensions

2. **`002_create_folders.sql`** - Creates folders table
   - Creates: `folders` table
   - Indexes: path, name

3. **`003_create_folders_categories.sql`** - Folder-category relationship
   - Creates: `folders_categories` junction table
   - Depends on: folders, categories

4. **`004_add_folder_to_files.sql`** - Link files to folders
   - Adds: `folder_id` column to `files`
   - Depends on: files, folders

5. **`005_add_parent_folder_id.sql`** - Folder hierarchy
   - Adds: `parent_folder_id` to `folders`
   - Enables: nested folder structure

6. **`006_add_soft_delete_to_categories.sql`** - Soft delete
   - Adds: `deleted_at` column to `categories`
   - Enables: soft delete for categories

## Running Migrations

### Using Makefile (recommended)
```bash
make migrate-up
```

### Manually with psql
```bash
# Apply all migrations in order
for migration in migrations/*.sql; do
    psql $DATABASE_URL -f "$migration"
done
```

### Individual migration
```bash
psql $DATABASE_URL -f migrations/001_init.sql
```

## CI/CD

Migrations are automatically applied in GitHub Actions from `migrations/` directory.
See `.github/workflows/test.yml` for details.

## Notes

- All migrations are in `migrations/` directory (no other locations)
- Migrations are numbered sequentially (001, 002, 003...)
- Applied in alphabetical/numerical order by `*.sql` glob
- Never modify existing migrations that have been applied to production
- Always create new migrations for schema changes
- Old location (`internal/db/migrations/`) is deprecated and should be ignored
