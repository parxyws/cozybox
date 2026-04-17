-- =============================================================================
-- Cozybox — Complete Schema
-- =============================================================================
-- Design decisions:
--   - ULID-style VARCHAR(26) primary keys (sortable, collision-resistant)
--   - sql.NullTime in Go → TIMESTAMPTZ in Postgres (ORM-portable)
--   - Soft deletes via deleted_at TIMESTAMPTZ (manual filter in queries)
--   - 1 Tenant = 1 Organization (UNIQUE constraint, unlockable later)
--   - All child tables carry tenant_id for direct scoped queries
--   - No accounts table (deferred — not in document generation scope)
--   - PDF stored in S3/MinIO; only the object key is stored in DB
-- =============================================================================


-- =============================================================================
-- IDENTITY LAYER
-- =============================================================================

CREATE TABLE users (
    id                    VARCHAR(26)  NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    username              VARCHAR(100) NOT NULL,
    email                 VARCHAR(255) NOT NULL,
    password              TEXT         NOT NULL DEFAULT '',
    -- Empty password = provisioned account, login blocked until setup is complete.
    -- Check: password = '' → return "account_setup_pending" error on login.

    status                VARCHAR(30)  NOT NULL DEFAULT 'active',
    -- 'active' | 'must_change_password' | 'suspended'

    force_password_change BOOLEAN      NOT NULL DEFAULT FALSE,
    -- TRUE on owner-provisioned accounts until member sets their own password.
    -- Triggers a restricted JWT scope ("setup_only") on login until cleared.

    is_verified           BOOLEAN      NOT NULL DEFAULT FALSE,
    -- FALSE until email OTP is confirmed. Self-registered users start FALSE.
    -- Owner-provisioned users start TRUE (owner vouches for the email).

    last_login            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uq_users_email    UNIQUE (email),
    CONSTRAINT uq_users_username UNIQUE (username)
);

CREATE INDEX idx_users_email      ON users(email)    WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username   ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status     ON users(status)   WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE account_setup_tokens (
    -- One-time setup links for owner-provisioned accounts.
    -- Owner creates member account → system sends email with a tokenized link.
    -- Member clicks link → sets their own password → token marked used → auto-login.
    -- Never stores the actual password. Token is the only credential.

    id          VARCHAR(26)  NOT NULL,
    user_id     VARCHAR(26)  NOT NULL,
    created_by  VARCHAR(26)  NOT NULL, -- the owner who provisioned this account
    token       VARCHAR(64)  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- 'pending' | 'used' | 'expired' | 'revoked'
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_account_setup_tokens     PRIMARY KEY (id),
    CONSTRAINT uq_account_setup_tokens_token UNIQUE (token),
    CONSTRAINT fk_setup_tokens_user        FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_setup_tokens_created_by  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_setup_tokens_token  ON account_setup_tokens(token);
CREATE INDEX idx_setup_tokens_user   ON account_setup_tokens(user_id);
CREATE INDEX idx_setup_tokens_status ON account_setup_tokens(status);


-- =============================================================================
-- TENANCY LAYER
-- =============================================================================

CREATE TABLE tenants (
    -- Top-level isolation boundary. Every piece of data is scoped to a tenant.
    -- Created during onboarding (single transaction with organization).
    -- 1 Tenant = 1 Organization enforced via UNIQUE on organizations(tenant_id).

    id         VARCHAR(26)  NOT NULL,
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL, -- URL-safe identifier, e.g. "acme-corp"
    status     VARCHAR(20)  NOT NULL DEFAULT 'active',
    -- 'active' | 'suspended' | 'cancelled'
    owner_id   VARCHAR(26)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_tenants      PRIMARY KEY (id),
    CONSTRAINT uq_tenants_slug UNIQUE (slug),
    CONSTRAINT fk_tenants_owner FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE INDEX idx_tenants_owner ON tenants(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_slug  ON tenants(slug)     WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE tenant_members (
    -- Junction table: maps users to tenants with a role.
    -- A user can belong to multiple tenants (freelancer, multi-company).
    -- The active tenant is stored in Redis session, not here.

    id        VARCHAR(26) NOT NULL,
    tenant_id VARCHAR(26) NOT NULL,
    user_id   VARCHAR(26) NOT NULL,
    role      VARCHAR(20) NOT NULL DEFAULT 'member',
    -- 'owner' | 'admin' | 'member'
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_tenant_members       PRIMARY KEY (id),
    CONSTRAINT uq_tenant_members_pair  UNIQUE (tenant_id, user_id),
    CONSTRAINT fk_tenant_members_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_tenant_members_user   FOREIGN KEY (user_id)   REFERENCES users(id)   ON DELETE CASCADE
);

CREATE INDEX idx_tenant_members_tenant ON tenant_members(tenant_id);
CREATE INDEX idx_tenant_members_user   ON tenant_members(user_id);

-- -----------------------------------------------------------------------------

CREATE TABLE tenant_invitations (
    -- Pending invitations sent by owner/admin to join a tenant.
    -- Used for the invite-link flow: owner sends email → recipient clicks link
    -- → registers (if new) or logs in (if existing) → accepts invitation.
    -- Separate from account_setup_tokens: invitations require user action to accept,
    -- setup tokens bypass acceptance and go straight to password setup.

    id          VARCHAR(26)  NOT NULL,
    tenant_id   VARCHAR(26)  NOT NULL,
    invited_by  VARCHAR(26)  NOT NULL,
    email       VARCHAR(255) NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'member',
    token       VARCHAR(64)  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- 'pending' | 'accepted' | 'declined' | 'expired' | 'revoked'
    expires_at  TIMESTAMPTZ  NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_tenant_invitations       PRIMARY KEY (id),
    CONSTRAINT uq_tenant_invitations_token UNIQUE (token),
    CONSTRAINT fk_invitations_tenant       FOREIGN KEY (tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_invitations_invited_by   FOREIGN KEY (invited_by) REFERENCES users(id)
);

CREATE INDEX idx_invitations_token  ON tenant_invitations(token);
CREATE INDEX idx_invitations_email  ON tenant_invitations(email);
CREATE INDEX idx_invitations_tenant ON tenant_invitations(tenant_id);
CREATE INDEX idx_invitations_status ON tenant_invitations(status);


-- =============================================================================
-- ORGANIZATION LAYER
-- =============================================================================

CREATE TABLE organizations (
    -- Business entity that owns documents and contacts.
    -- Auto-created in the same transaction as the tenant (onboarding).
    -- UNIQUE(tenant_id) enforces 1 tenant = 1 org for now.
    -- Drop the constraint if multi-org is ever unlocked.

    id               VARCHAR(26)  NOT NULL,
    tenant_id        VARCHAR(26)  NOT NULL, -- UNIQUE: 1 tenant = 1 org
    owner_id         VARCHAR(26)  NOT NULL,
    name             VARCHAR(255) NOT NULL,
    email            VARCHAR(255),
    phone            VARCHAR(50),
    website          VARCHAR(255),
    address_line1    VARCHAR(255),
    address_line2    VARCHAR(255),
    city             VARCHAR(100),
    state            VARCHAR(100),
    postal_code      VARCHAR(20),
    country          VARCHAR(10),
    tax_id           VARCHAR(100),
    logo_s3_key      TEXT,
    -- S3/MinIO object key for the org logo. Generate presigned URL at request time.
    -- Never store the presigned URL itself — it expires.
    default_currency VARCHAR(10)  NOT NULL DEFAULT 'USD',
    timezone         VARCHAR(50)  NOT NULL DEFAULT 'UTC',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,

    CONSTRAINT pk_organizations        PRIMARY KEY (id),
    CONSTRAINT uq_organizations_tenant UNIQUE (tenant_id),  -- 1:1 with tenant
    CONSTRAINT fk_organizations_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_organizations_owner  FOREIGN KEY (owner_id)  REFERENCES users(id)
);

CREATE INDEX idx_organizations_tenant ON organizations(tenant_id) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE contacts (
    -- Clients, suppliers, or both — linked to the organization.
    -- Used as recipients on documents (invoices sent to clients,
    -- purchase orders sent to suppliers).

    id              VARCHAR(26)  NOT NULL,
    tenant_id       VARCHAR(26)  NOT NULL, -- denormalized for direct scoped queries
    organization_id VARCHAR(26)  NOT NULL,
    type            VARCHAR(20)  NOT NULL DEFAULT 'client',
    -- 'client' | 'supplier' | 'both'
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255),
    phone           VARCHAR(50),
    address_line1   VARCHAR(255),
    address_line2   VARCHAR(255),
    city            VARCHAR(100),
    state           VARCHAR(100),
    postal_code     VARCHAR(20),
    country         VARCHAR(10),
    tax_id          VARCHAR(100),
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT pk_contacts        PRIMARY KEY (id),
    CONSTRAINT fk_contacts_tenant FOREIGN KEY (tenant_id)       REFERENCES tenants(id),
    CONSTRAINT fk_contacts_org    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_contacts_tenant ON contacts(tenant_id)       WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_org    ON contacts(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_type   ON contacts(type)            WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE document_sequences (
    -- Auto-numbering config per organization per document type.
    -- Generates refs like: INV/2026/04/0001, QUO/2026/04/0002.
    -- Auto-created (one row per doc type) when the organization is created.
    -- Increment via SELECT ... FOR UPDATE in a transaction (not Redis).
    -- Redis can be used as a read-ahead cache, but Postgres is the source of truth.

    id              VARCHAR(26)  NOT NULL,
    tenant_id       VARCHAR(26)  NOT NULL,
    organization_id VARCHAR(26)  NOT NULL,
    type            VARCHAR(30)  NOT NULL,
    -- 'quotation' | 'invoice' | 'receipt' | 'purchase_order' | 'sales_order' | 'debit_note'
    prefix          VARCHAR(20)  NOT NULL,
    -- e.g. 'INV', 'QUO', 'REC', 'PO', 'SO', 'DN'
    next_number     INT          NOT NULL DEFAULT 1,
    format          VARCHAR(100) NOT NULL DEFAULT '{PREFIX}/{YEAR}/{MONTH}/{SEQ:4}',
    last_reset_at   TIMESTAMPTZ,
    -- Tracks when next_number was last reset (monthly/annual reset support)
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_sequences        PRIMARY KEY (id),
    CONSTRAINT uq_document_sequences_type   UNIQUE (organization_id, type),
    CONSTRAINT fk_sequences_tenant          FOREIGN KEY (tenant_id)       REFERENCES tenants(id),
    CONSTRAINT fk_sequences_org             FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_sequences_tenant ON document_sequences(tenant_id);
CREATE INDEX idx_sequences_org    ON document_sequences(organization_id);


-- =============================================================================
-- DOCUMENT LAYER
-- =============================================================================

CREATE TABLE documents (
    -- Unified table for all 6 document types.
    -- Type discriminator column determines shape; JSONB metadata handles type-specific fields.
    -- Document chain: quotation → invoice → receipt (parent_id tracks the lineage).
    -- Status transitions are enforced in the service layer (state machine), not here.

    id              VARCHAR(26)    NOT NULL,
    tenant_id       VARCHAR(26)    NOT NULL,
    organization_id VARCHAR(26)    NOT NULL,
    contact_id      VARCHAR(26),   -- nullable: draft may not have a contact yet
    parent_id       VARCHAR(26),   -- nullable: root documents have no parent
    created_by      VARCHAR(26)    NOT NULL,

    -- Classification
    type            VARCHAR(30)    NOT NULL,
    -- 'quotation' | 'invoice' | 'receipt' | 'purchase_order' | 'sales_order' | 'debit_note'

    status          VARCHAR(30)    NOT NULL DEFAULT 'draft',
    -- 'draft' → 'published' → 'accepted' | 'rejected' | 'overdue'
    --         → 'paid' | 'partially_paid'
    -- Any non-paid status → 'cancelled'

    document_ref    VARCHAR(100)   NOT NULL DEFAULT '',
    -- Assigned at publish, not at create. Empty on drafts.
    -- Format: INV/2026/04/0001 (generated from document_sequences)

    -- Dates (TIMESTAMPTZ: time component always 00:00:00 UTC for calendar dates)
    issue_date      TIMESTAMPTZ,
    due_date        TIMESTAMPTZ,   -- triggers overdue check if unpaid past this date
    valid_until     TIMESTAMPTZ,   -- quotation expiry

    -- Financial totals (computed from document_items, stored for fast queries)
    currency        VARCHAR(10)    NOT NULL DEFAULT 'USD',
    subtotal        NUMERIC(15, 4) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    tax_amount      NUMERIC(15, 4) NOT NULL DEFAULT 0,
    total           NUMERIC(15, 4) NOT NULL DEFAULT 0,
    amount_paid     NUMERIC(15, 4) NOT NULL DEFAULT 0,
    -- amount_paid updated via partial-pay / pay status transitions

    -- Free-text content
    notes           TEXT,
    terms           TEXT,
    footer          TEXT,

    -- Type-specific fields as JSONB (avoid ALTER TABLE for every doc type quirk)
    metadata        JSONB          NOT NULL DEFAULT '{}',

    -- PDF storage: S3/MinIO object key (not URL — URLs expire)
    pdf_s3_key      TEXT,
    -- pdfs/{tenant_id}/{org_id}/{type}/{year}/{month}/{doc_ref}.pdf
    pdf_generated_at TIMESTAMPTZ,
    pdf_size_bytes  INT,

    -- Audit
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT pk_documents         PRIMARY KEY (id),
    CONSTRAINT fk_documents_tenant  FOREIGN KEY (tenant_id)       REFERENCES tenants(id),
    CONSTRAINT fk_documents_org     FOREIGN KEY (organization_id) REFERENCES organizations(id),
    CONSTRAINT fk_documents_contact FOREIGN KEY (contact_id)      REFERENCES contacts(id),
    CONSTRAINT fk_documents_parent  FOREIGN KEY (parent_id)       REFERENCES documents(id),
    CONSTRAINT fk_documents_creator FOREIGN KEY (created_by)      REFERENCES users(id)
);

CREATE INDEX idx_documents_tenant   ON documents(tenant_id)       WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_org      ON documents(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_contact  ON documents(contact_id)      WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_parent   ON documents(parent_id)       WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_type     ON documents(type)            WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_status   ON documents(status)          WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_ref      ON documents(document_ref)    WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_due_date ON documents(due_date)        WHERE deleted_at IS NULL;
-- idx_documents_due_date: used by scheduled job to find overdue documents

-- -----------------------------------------------------------------------------

CREATE TABLE document_items (
    -- Line items belonging to a document.
    -- Deleted with the document (CASCADE). No soft delete needed here.
    -- amount = (quantity × unit_price) - (discount_pct%) + (tax_pct%)
    -- Recomputed and stored on every save; document totals are the sum.

    id           VARCHAR(26)    NOT NULL,
    document_id  VARCHAR(26)    NOT NULL,
    sort_order   INT            NOT NULL DEFAULT 0,
    description  TEXT           NOT NULL,
    quantity     NUMERIC(15, 4) NOT NULL DEFAULT 1,
    unit         VARCHAR(50),   -- e.g. 'pcs', 'hrs', 'kg'
    unit_price   NUMERIC(15, 4) NOT NULL DEFAULT 0,
    discount_pct NUMERIC(5, 2)  NOT NULL DEFAULT 0,  -- percentage, e.g. 10.00 = 10%
    tax_pct      NUMERIC(5, 2)  NOT NULL DEFAULT 0,  -- percentage, e.g. 11.00 = 11% PPN
    amount       NUMERIC(15, 4) NOT NULL DEFAULT 0,  -- computed, stored
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_items     PRIMARY KEY (id),
    CONSTRAINT fk_items_document     FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE INDEX idx_document_items_doc ON document_items(document_id);

-- -----------------------------------------------------------------------------

CREATE TABLE document_activities (
    -- Append-only audit log. Never update or delete rows here.
    -- Records every lifecycle event: creation, edits, status transitions, PDF generation, email sends.
    -- performed_by is NULL for system-triggered actions (e.g. overdue scheduled job).

    id           VARCHAR(26) NOT NULL,
    document_id  VARCHAR(26) NOT NULL,
    performed_by VARCHAR(26),
    action       VARCHAR(50) NOT NULL,
    -- 'created' | 'updated' | 'status_changed' | 'pdf_generated' | 'sent' | 'payment_recorded'
    from_status  VARCHAR(30),
    to_status    VARCHAR(30),
    note         TEXT,
    metadata     JSONB       NOT NULL DEFAULT '{}',
    -- Snapshot of what changed. Examples:
    --   status_changed:  { "old": "draft", "new": "published" }
    --   updated:         { "changed_fields": ["due_date", "total"], "old": {...}, "new": {...} }
    --   sent:            { "to": "client@example.com", "subject": "Invoice INV/2026/04/0001" }
    --   pdf_generated:   { "s3_key": "pdfs/...", "size_bytes": 84210 }
    --   payment_recorded:{ "amount": "5000000.00", "method": "bank_transfer" }
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_activities     PRIMARY KEY (id),
    CONSTRAINT fk_activities_document     FOREIGN KEY (document_id)  REFERENCES documents(id) ON DELETE CASCADE,
    CONSTRAINT fk_activities_performed_by FOREIGN KEY (performed_by) REFERENCES users(id)
);

CREATE INDEX idx_activities_document ON document_activities(document_id);
CREATE INDEX idx_activities_actor    ON document_activities(performed_by);
CREATE INDEX idx_activities_action   ON document_activities(action);
