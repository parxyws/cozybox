-- =============================================
-- Rollback: Remove multi-tenancy support
-- =============================================

-- Drop indexes first
DROP INDEX IF EXISTS idx_document_sequences_tenant;
DROP INDEX IF EXISTS idx_contacts_tenant;
DROP INDEX IF EXISTS idx_documents_tenant;
DROP INDEX IF EXISTS idx_organizations_tenant;
DROP INDEX IF EXISTS idx_tenant_members_tenant;
DROP INDEX IF EXISTS idx_tenant_members_user;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP INDEX IF EXISTS idx_tenants_owner;

-- Remove tenant_id from existing tables
ALTER TABLE document_sequences DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE contacts DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS tenant_id;

-- Drop tenant tables (order matters for FK constraints)
DROP TABLE IF EXISTS tenant_members;
DROP TABLE IF EXISTS tenants;
