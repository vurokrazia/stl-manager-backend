# Database Migrations

This directory contains incremental database migrations.

## Migration Order

Migrations are applied in the following order:

1. **`internal/db/migrations/001_init.sql`** (FIRST - creates base tables)
   - Creates initial tables: `scans`, `files`, `categories`, `files_categories`
   - Required before all other migrations

2. **`migrations/004_create_folders.sql`**
   - Creates `folders` table

3. **`migrations/005_create_folders_categories.sql`**
   - Creates `folders_categories` junction table
   - Depends on: folders, categories

4. **`migrations/006_add_folder_to_files.sql`**
   - Adds `folder_id` column to `files` table
   - Depends on: files, folders

5. **`migrations/007_add_parent_folder_id.sql`**
   - Adds `parent_folder_id` to `folders` table for hierarchy
   - Depends on: folders

6. **`migrations/008_add_soft_delete_to_categories.sql`**
   - Adds `deleted_at` column to `categories` table
   - Depends on: categories

## Running Migrations

### Using Makefile (recommended)
```bash
make migrate-up
```

### Manually with psql
```bash
# 1. Apply initial migration
psql $DATABASE_URL -f internal/db/migrations/001_init.sql

# 2. Apply incremental migrations
for migration in migrations/*.sql; do
    psql $DATABASE_URL -f "$migration"
done
```

## CI/CD

Migrations are automatically applied in GitHub Actions in the correct order.
See `.github/workflows/test.yml` for details.

## Notes

- The initial migration (001_init.sql) is in a separate directory (`internal/db/migrations/`)
- Incremental migrations (004-008) are numbered to maintain order
- Never modify existing migrations that have been applied to production
- Always create new migrations for schema changes
