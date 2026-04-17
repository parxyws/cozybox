-- =============================================================================
-- Cozybox — Drop Complete Schema (Rollback)
-- =============================================================================
-- Drop order respects FK dependencies: children before parents.
-- =============================================================================

-- Document layer
DROP TABLE IF EXISTS document_activities;
DROP TABLE IF EXISTS document_items;
DROP TABLE IF EXISTS documents;

-- Organization layer
DROP TABLE IF EXISTS document_sequences;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS organizations;

-- Tenancy layer
DROP TABLE IF EXISTS tenant_invitations;
DROP TABLE IF EXISTS tenant_members;
DROP TABLE IF EXISTS tenants;

-- Identity layer
DROP TABLE IF EXISTS account_setup_tokens;
DROP TABLE IF EXISTS users;
