-- =============================================
-- Migration: Add multi-tenancy support
-- =============================================

-- Tenant: the top-level isolation boundary
CREATE TABLE tenants (
    id              VARCHAR(26) PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    owner_id        VARCHAR(26) NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- TenantMember: maps users to tenants with roles
CREATE TABLE tenant_members (
    id              VARCHAR(26) PRIMARY KEY,
    tenant_id       VARCHAR(26) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         VARCHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id)
);

-- Add tenant_id to existing tables
ALTER TABLE organizations ADD COLUMN tenant_id VARCHAR(26) REFERENCES tenants(id);
ALTER TABLE documents ADD COLUMN tenant_id VARCHAR(26) REFERENCES tenants(id);
ALTER TABLE contacts ADD COLUMN tenant_id VARCHAR(26) REFERENCES tenants(id);
ALTER TABLE document_sequences ADD COLUMN tenant_id VARCHAR(26) REFERENCES tenants(id);

-- Indexes for tenant-scoped queries
CREATE INDEX idx_tenants_owner ON tenants(owner_id);
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX idx_tenant_members_tenant ON tenant_members(tenant_id);
CREATE INDEX idx_organizations_tenant ON organizations(tenant_id);
CREATE INDEX idx_documents_tenant ON documents(tenant_id);
CREATE INDEX idx_contacts_tenant ON contacts(tenant_id);
CREATE INDEX idx_document_sequences_tenant ON document_sequences(tenant_id);
