-- 000001_create_tables.down.sql
-- Drop tables in reverse order (profiles first due to foreign key)
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS tenants;
