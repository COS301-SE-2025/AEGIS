-- =============================
-- RBAC + Case User Roles Schema (ENUM-based, simplified)
-- =============================


--****************************
-- 1.  docker cp ./schema.sql aegis-postgres-1:/schema.sql

-- 2. docker exec -it aegis-postgres-1 psql -U app_user -d app_database -f /schema.sql

--****************************
-- Create ENUM types
CREATE TYPE user_role AS ENUM (
    -- 🔐 Core Operational & Admin Roles
    'Admin', 
    'System Admin',
    'DFIR Admin',  
    'Tenant Admin',                  -- Manages tenant-level operations, teams, and users                         
    'External Collaborator',                      -- Basic access, limited permissions

    -- 🛡️ Incident Response & Forensics Roles
    'Incident Responder',                -- Handles containment, eradication, recovery
    'Forensic Analyst',                 -- Analyzes digital evidence to determine attack scope
    'Malware Analyst',                  -- Dissects malicious software to understand behavior
    'Threat Intelligent Analyst',       -- Correlates data with threat actor TTPs and threat feeds
    'DFIR Manager',                     -- Leads DFIR efforts, manages team response and priorities
    'Detection Engineer',               -- Develops and tunes detection rules and alert systems

    -- 🧪 Specialized Evidence Analysis Roles
    'Network Evidence Analyst',         -- Focuses on packet captures, flow data, firewall logs
    'Image Forensics Analyst',          -- Examines images for tampering, EXIF data, hidden content
    'Disk Forensics Analyst',           -- Analyzes disk images for deleted files, artifacts
    'Log Analyst',                      -- Parses system and application logs to find indicators
    'Mobile Device Analyst',            -- Extracts and examines data from smartphones/tablets
    'Memory Forensics Analyst',         -- Investigates volatile memory dumps for malware/residue
    'Cloud Forensics Specialist',       -- Investigates cloud-hosted evidence, e.g., AWS, Azure
    'Endpoint Forensics Analyst',       -- Analyzes compromised endpoint systems and artifacts
    'Reverse Engineer',                 -- Disassembles malware and binaries to identify functionality
    'SIEM Analyst',                     -- Monitors alerts and correlates events in SIEM platforms
    'Vulnerability Analyst',            -- Identifies vulnerabilities and assists in risk assessment
    'Digital Evidence Technician',      -- Supports imaging, hashing, chain-of-custody handling
    'Packet Analyst',                   -- Deep analysis of network packets and PCAP files

    -- ⚖️ Legal, Compliance & Governance Roles
    'Legal/Compliance Liaison',         -- Coordinates with legal/compliance during investigations
    'Compliance Officer',               -- Ensures response actions follow regulations (e.g., POPIA, GDPR)
    'Legal Counsel',                    -- Provides legal guidance on investigations and evidence handling
    'Policy Analyst',                   -- Reviews and drafts security and investigation policies

    -- 📢 Operational & Coordination Roles
    'SOC Analyst',                      -- Tier 1–3 responder within a Security Operations Center
    'Incident Commander',               -- Manages major incidents, makes executive decisions
    'Crisis Communications Officer',    -- Handles public and internal messaging during incidents
    'IT Infrastructure Liaison',        -- Coordinates technical access, remediation with IT teams
    'Triage Analyst',                   -- Prioritizes and classifies incoming alerts and reports

    -- 🗃️ Supportive & Oversight Roles
    'Evidence Archivist',               -- Maintains long-term storage and retrieval of digital evidence
    'Training Coordinator',             -- Provides training and onboarding for DFIR tools and policy
    'Audit Reviewer',                   -- Reviews logs and user actions for audit and compliance
    'Threat Hunter'                     -- Proactively searches for hidden or advanced threats
);

CREATE TYPE case_status AS ENUM ('open', 'under_review', 'closed','ongoing','archived');
CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'critical','time-sensitive');
CREATE TYPE investigation_stage AS ENUM (
    'Triage',
    'Evidence Collection',
    'Analysis',
    'Correlation & Threat Intelligence',
    'Containment & Eradication',
    'Recovery',
    'Reporting & Documentation',
    'Case Closure & Review'
);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  team_name TEXT NOT NULL,
  tenant_id UUID NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_teams_tenant FOREIGN KEY (tenant_id)
    REFERENCES tenants(id) ON DELETE CASCADE
);
-- STEP 1: Create ENUM for token status (if it doesn't exist yet)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'token_status') THEN
        CREATE TYPE token_status AS ENUM ('active', 'expired', 'revoked');
    END IF;
END$$;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role,
    is_verified BOOLEAN DEFAULT FALSE,
    
    -- New Fields
    profile_picture_url TEXT,
    token_version INTEGER DEFAULT 1, -- for JWT version control
    external_token_expiry TIMESTAMP, -- expiry for external collaborator token
    external_token_status token_status DEFAULT 'active', -- token status: active, expired, revoked
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    email_verified_at TIMESTAMP NULL,
    accepted_terms_at TIMESTAMP,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL, -- Link to tenant
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL -- Link to team
);

CREATE TYPE token_type AS ENUM (
    'EMAIL_VERIFY',
    'RESET_PASSWORD',
    'COLLAB_INVITE'
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    read BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE
);





CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type token_type NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    used BOOLEAN DEFAULT FALSE
);

-- X3DH Identity Keys table
CREATE TABLE x3dh_identity_keys (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,  -- Encrypt this field before storing
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE x3dh_signed_prekeys (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,   -- Encrypted
    signature TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    COLUMN expires_at TIMESTAMPTZ,
    expires_at TIMESTAMP
);


CREATE TABLE x3dh_one_time_prekeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,   -- Encrypted
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Cases table
CREATE TABLE IF NOT EXISTS cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    status case_status DEFAULT 'open',
    investigation_stage investigation_stage DEFAULT 'Triage',
    priority case_priority DEFAULT 'medium',
    team_name TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- New Fields
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL, -- Link to team
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

-- Groups table
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,

    case_id UUID REFERENCES cases(id) ON DELETE CASCADE,
    group_url TEXT, -- emoji or image URL for avatar

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- New Fields
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

CREATE TABLE annotation_threads (
  id UUID PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  file_id UUID NOT NULL,
  case_id UUID NOT NULL,
  created_by UUID NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  status VARCHAR(50) DEFAULT 'open',
  priority VARCHAR(50) DEFAULT 'medium',
  is_active BOOLEAN DEFAULT true,
  resolved_at TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
    -- New Fields
);

CREATE TABLE thread_tags (
  id UUID PRIMARY KEY,
  thread_id UUID NOT NULL REFERENCES annotation_threads(id),
  tag_name VARCHAR(255) NOT NULL
);

CREATE TABLE thread_participants (
  thread_id UUID NOT NULL REFERENCES annotation_threads(id),
  user_id UUID NOT NULL,
  joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (thread_id, user_id)
);

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Now create your table
CREATE TABLE thread_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID NOT NULL,
    parent_message_id UUID,
    user_id UUID NOT NULL,
    message TEXT NOT NULL,
    is_approved BOOLEAN,
    approved_by UUID,
    approved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- New Fields
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

CREATE TABLE message_mentions (
    message_id UUID NOT NULL,
    mentioned_user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, mentioned_user_id),
    -- New Fields
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);
CREATE TABLE message_reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL,
    user_id UUID NOT NULL,
    reaction TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- New Fields   
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

CREATE TABLE threads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

ALTER TABLE thread_messages
  ADD FOREIGN KEY (parent_message_id) REFERENCES thread_messages(id),
  ADD FOREIGN KEY (thread_id) REFERENCES threads(id),
  ADD FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE message_mentions
  ADD FOREIGN KEY (message_id) REFERENCES thread_messages(id),
  ADD FOREIGN KEY (mentioned_user_id) REFERENCES users(id);

ALTER TABLE message_reactions
  ADD FOREIGN KEY (message_id) REFERENCES thread_messages(id),
  ADD FOREIGN KEY (user_id) REFERENCES users(id);

-- Evidence table
CREATE TABLE IF NOT EXISTS evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL,
    ipfs_cid TEXT NOT NULL,
    file_size INTEGER CHECK (file_size >= 0),
    checksum TEXT NOT NULL,
    metadata JSONB, -- stores metadata as a key-value JSON object
    uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    -- New Fields
    -- 🔹 Multi-tenancy fields
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    team_id   UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE
);



-- Optional indexes for performance
-- Index for searching by case
CREATE INDEX idx_evidence_case_id ON evidence(case_id);

-- Index for searching by tenant/team
CREATE INDEX idx_evidence_tenant_id ON evidence(tenant_id);
CREATE INDEX idx_evidence_team_id   ON evidence(team_id);

-- Index for searching by checksum (fast duplicate detection)
CREATE INDEX idx_evidence_checksum  ON evidence(checksum);

-- Tags and linking table
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS evidence_tags (
    evidence_id UUID REFERENCES evidence(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (evidence_id, tag_id)
);






--****************************
---------------------------

------------------------
--****************************--
--Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL
);

-- Insert: User Management
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'user:create', 'Create/register a new user'),
-- (gen_random_uuid(), 'user:view', 'View user details'),
-- (gen_random_uuid(), 'user:update', 'Edit user details'),
-- (gen_random_uuid(), 'user:delete', 'Remove a user from the system'),
-- (gen_random_uuid(), 'user:assign_role', 'Assign roles to users'),
-- (gen_random_uuid(), 'user:verify', 'Manually verify a user''s account'),
-- (gen_random_uuid(), 'user:reset_password', 'Initiate password reset for a user');

-- Insert: Evidence Management
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'evidence:upload', 'Upload new evidence file'),
-- (gen_random_uuid(), 'evidence:view', 'View/download evidence'),
-- (gen_random_uuid(), 'evidence:delete', 'Delete evidence (soft/hard delete)'),
-- (gen_random_uuid(), 'evidence:update_metadata', 'Edit metadata fields for evidence'),
-- (gen_random_uuid(), 'evidence:tag', 'Assign tags to evidence'),
-- (gen_random_uuid(), 'evidence:assign_analyst', 'Assign an analyst to a specific evidence item'),
-- (gen_random_uuid(), 'evidence:verify_checksum', 'Validate integrity of uploaded file'),
-- (gen_random_uuid(), 'evidence:view_logs', 'View audit logs related to evidence actions');

-- Extend Evidence Tagging
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'evidence:create_tag', 'Create new tags for evidence'),
-- (gen_random_uuid(), 'evidence:remove_tag', 'Remove existing tags from evidence');

-- Insert: Specialized Evidence Analysis
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'analysis:network', 'Perform network traffic analysis'),
-- (gen_random_uuid(), 'analysis:image', 'Analyze digital images for forensics'),
-- (gen_random_uuid(), 'analysis:disk', 'Perform disk image analysis'),
-- (gen_random_uuid(), 'analysis:logs', 'Analyze logs and detect anomalies'),
-- (gen_random_uuid(), 'analysis:mobile', 'Analyze data from mobile devices'),
-- (gen_random_uuid(), 'analysis:memory', 'Conduct memory dump analysis'),
-- (gen_random_uuid(), 'analysis:cloud', 'Review cloud-stored forensic data'),
-- (gen_random_uuid(), 'analysis:endpoint', 'Investigate endpoint artifacts'),
-- (gen_random_uuid(), 'analysis:reverse_engineer', 'Reverse engineer a binary/malware sample');

-- Insert: Case Management
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'case:create', 'Create a new case'),
-- (gen_random_uuid(), 'case:view', 'View case details'),
-- (gen_random_uuid(), 'case:update', 'Update case information'),
-- (gen_random_uuid(), 'case:assign_user', 'Assign user to case'),
-- (gen_random_uuid(), 'case:change_status', 'Change case status (open, review, closed)'),
-- (gen_random_uuid(), 'case:set_priority', 'Set or modify case priority'),
-- (gen_random_uuid(), 'case:archive', 'Archive or lock case for compliance purposes');

-- Case Tagging
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'case:tag', 'Assign tags to cases'),
-- (gen_random_uuid(), 'case:create_tag', 'Create new tags for cases'),
-- (gen_random_uuid(), 'case:remove_tag', 'Remove existing tags from cases');


-- Insert: Dashboard & Audit
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'dashboard:view', 'View dashboards with case/evidence summaries'),
-- (gen_random_uuid(), 'audit:view_all', 'View all system-level audit logs'),
-- (gen_random_uuid(), 'audit:view_case', 'View audit logs related to a specific case'),
-- (gen_random_uuid(), 'audit:view_user', 'View audit actions performed by a user'),
-- (gen_random_uuid(), 'audit:generate_report', 'Export/print audit or investigation report');

-- Insert: Compliance & Legal
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'compliance:review', 'View compliance indicators for cases'),
-- (gen_random_uuid(), 'compliance:lock_evidence', 'Lock evidence to preserve legal integrity'),
-- (gen_random_uuid(), 'compliance:export_data', 'Export data for legal or regulatory use'),
-- (gen_random_uuid(), 'legal:comment', 'Add legal notes or directives to case/evidence'),
-- (gen_random_uuid(), 'legal:mark_sensitive', 'Mark evidence or case as sensitive/privileged');


-- Log Exporting
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'audit:export_case_log', 'Export audit logs for a specific case'),
-- (gen_random_uuid(), 'audit:export_user_log', 'Export audit logs for a specific user');

-- Annotation and Commenting
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'comment:create', 'Add a comment to a case or evidence'),
-- (gen_random_uuid(), 'comment:reply', 'Reply to an existing comment thread'),
-- (gen_random_uuid(), 'comment:edit', 'Edit your own comment'),
-- (gen_random_uuid(), 'comment:delete', 'Delete a comment you authored'),
-- (gen_random_uuid(), 'comment:moderate', 'Moderate or remove inappropriate comments'),
-- (gen_random_uuid(), 'thread:create', 'Start a new annotation thread on a case or file'),
-- (gen_random_uuid(), 'thread:close', 'Close or archive a thread to prevent new replies');

-- Case Collaboration
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'collaboration:add_member', 'Add member to a case collaboration team'),
-- (gen_random_uuid(), 'collaboration:remove_member', 'Remove member from a case team'),
-- (gen_random_uuid(), 'collaboration:change_role', 'Update user role within a case'),
-- (gen_random_uuid(), 'collaboration:message', 'Send secure messages within a case context');

-- -- Secure Chat
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'chat:create_room', 'Create a new chat room for a case'),
-- (gen_random_uuid(), 'chat:send_message', 'Send a message in a secure chat'),
-- (gen_random_uuid(), 'chat:view_history', 'View chat history'),
-- (gen_random_uuid(), 'chat:delete_message', 'Delete your message from chat'),
-- (gen_random_uuid(), 'chat:moderate', 'Moderate chat content or users');

-- Maps ENUM roles to permissions directly
CREATE TABLE enum_role_permissions (
    role user_role NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);

-- Optional: global user roles if needed alongside enum
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- Per-case user role assignment using ENUM
CREATE TABLE case_user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    role user_role NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, case_id),
    -- New Fields
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL -- Link to tenant
);

-- Indexes to support performance (optional but recommended)
CREATE INDEX idx_case_user_roles_case_id ON case_user_roles(case_id);
CREATE INDEX idx_case_user_roles_user_id ON case_user_roles(user_id);

CREATE TABLE entries (
    id VARCHAR(255) PRIMARY KEY,
    case_id VARCHAR(255) NOT NULL,
    evidence_id VARCHAR(255) NOT NULL,
    actor_id VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    reason TEXT,
    location TEXT,
    hash_md5 VARCHAR(32),
    hash_sha1 VARCHAR(40),
    hash_sha256 VARCHAR(64),
    occurred_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    
 
);

-- === Chain of Custody (AEGIS) Final Schema — Updated ===
-- Assumes you already have: cases(id), evidence(id), users(id)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1) Actions we actually support now
DO $$ BEGIN
  CREATE TYPE coc_action AS ENUM (
    'upload',    -- evidence first inserted into the system
    'download',  -- evidence retrieved from the system
    'archive',   -- evidence moved to long-term storage
    'view'       -- evidence opened/viewed in Evidence Viewer
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- 2) Main CoC table
CREATE TABLE IF NOT EXISTS chain_of_custody (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  case_id      UUID NOT NULL REFERENCES cases(id)    ON DELETE CASCADE,
  evidence_id  UUID NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
  actor_id     UUID              REFERENCES users(id), -- who performed the action (nullable for system)

  action       coc_action NOT NULL,
---  reason       TEXT,           -- justification / notes (optional)

  location     TEXT,           -- physical/logical location (optional)
  hash_md5     TEXT,           -- legacy compatibility (optional)
  hash_sha1    TEXT,           -- legacy compatibility (optional)
  hash_sha256  TEXT,           -- canonical integrity hash

  occurred_at  TIMESTAMPTZ NOT NULL,                -- when the action happened
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()   -- when we recorded it
);

-- 3) Helpful indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_coc_case_time
  ON chain_of_custody (case_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_coc_evidence_time
  ON chain_of_custody (evidence_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_coc_actor
  ON chain_of_custody (actor_id);

CREATE INDEX IF NOT EXISTS idx_coc_action
  ON chain_of_custody (action);

-- 4) Append-only enforcement
CREATE OR REPLACE FUNCTION forbid_coc_update_delete()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'Chain of Custody entries are immutable';
END $$;

CREATE TRIGGER trg_coc_no_update
BEFORE UPDATE ON chain_of_custody
FOR EACH ROW EXECUTE FUNCTION forbid_coc_update_delete();

CREATE TRIGGER trg_coc_no_delete
BEFORE DELETE ON chain_of_custody
FOR EACH ROW EXECUTE FUNCTION forbid_coc_update_delete();

-- 5) Convenience view with actor data for UI
CREATE OR REPLACE VIEW v_chain_of_custody_with_actor AS
SELECT
  c.id,
  c.case_id,
  c.evidence_id,
  c.actor_id,
  u.name  AS actor_name,
  u.email AS actor_email,
  c.action,
  c.reason,
  c.location,
  c.hash_md5,
  c.hash_sha1,
  c.hash_sha256,
  c.occurred_at,
  c.created_at
FROM chain_of_custody c
LEFT JOIN users u ON u.id = c.actor_id
ORDER BY c.occurred_at ASC, c.created_at ASC;



-- For reference (no change needed if this already exists)
CREATE TABLE IF NOT EXISTS report_hashes (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  device_id UUID NOT NULL REFERENCES report_devices(id) ON DELETE CASCADE,
  md5       TEXT,
  sha1      TEXT,
  sha256    TEXT,
  context   TEXT,  -- e.g., "acquisition image", "working copy"
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- ALTER TABLE case_reports
--   ADD COLUMN IF NOT EXISTS acquisition_methods  TEXT,  -- how imaging/collection was done
--   ADD COLUMN IF NOT EXISTS analysis_techniques  TEXT;  -- e.g., keyword search, timeline analysis


DO $$ BEGIN
  CREATE TYPE report_status  AS ENUM ('draft', 'published', 'review', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE report_format  AS ENUM ('pdf','json','csv');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS case_reports (
  id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id                   UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  examiner_id               UUID NOT NULL REFERENCES users(id),
  
  -- Name and file path fields for report metadata
  name                      VARCHAR(255) NOT NULL,  -- Name of the report
  file_path                 VARCHAR(255) NOT NULL,  -- Path to the report file
  
  -- Narrative fields for the report's content
  scope                     TEXT,
  objectives                TEXT,
  limitations               TEXT,
  tools_methods             TEXT,   -- Simplified: single text field
  final_conclusion          TEXT,
  evidence_summary          TEXT,
  certification_statement   TEXT,
  date_examined             DATE,

  -- Lifecycle / admin fields
  status                    report_status NOT NULL DEFAULT 'draft',
  version                   INTEGER NOT NULL DEFAULT 1,
  report_number             TEXT UNIQUE,   -- Optional human-friendly ID
  
  -- Timestamps
  created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_case_reports_case_id     ON case_reports(case_id);
CREATE INDEX IF NOT EXISTS idx_case_reports_examiner_id ON case_reports(examiner_id);
CREATE INDEX IF NOT EXISTS idx_case_reports_status      ON case_reports(status);
-- 1) Create the helper function (idempotent)
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_case_reports_updated_at
BEFORE UPDATE ON case_reports
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- First, create the report_status enum type (PostgreSQL)
-- Enable UUID generation (PostgreSQL)
-- Enable UUID generation (PostgreSQL)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum type for report status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'report_status') THEN
        CREATE TYPE report_status AS ENUM ('draft', 'review', 'published');
    END IF;
END
$$;

-- Create the reports table
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id UUID NOT NULL,                  -- Links report to a specific case
    examiner_id UUID NOT NULL,              -- Who created the report
    name VARCHAR(255) NOT NULL,             -- Report title
    report_number VARCHAR(255) UNIQUE,      -- Optional, unique identifier
    version INTEGER NOT NULL DEFAULT 1,     -- Versioning
    status report_status DEFAULT 'draft',   -- Current report status
    scope TEXT,                             -- Optional field for report scope
    objectives TEXT,                        -- Optional objectives
    limitations TEXT,                       -- Optional limitations
    tools_methods TEXT,                      -- Optional tools & methods
    evidence_summary TEXT,                  -- Optional summary of evidence
    final_conclusion TEXT,                  -- Optional conclusions
    certification_statement TEXT,           -- Optional certification / sign-off
    date_examined DATE,                      -- Optional examination date
    file_path VARCHAR(255) NOT NULL,        -- Path to stored PDF or file
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better query performance
CREATE INDEX idx_reports_case_id ON reports(case_id);
CREATE INDEX idx_reports_examiner_id ON reports(examiner_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_created_at ON reports(created_at);
CREATE INDEX idx_reports_report_number ON reports(report_number);
CREATE INDEX idx_reports_date_examined ON reports(date_examined);

-- Add foreign key constraints (uncomment if you have related tables)
-- ALTER TABLE reports ADD CONSTRAINT fk_reports_case_id 
--     FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE;
-- ALTER TABLE reports ADD CONSTRAINT fk_reports_examiner_id 
--     FOREIGN KEY (examiner_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_reports_updated_at 
    BEFORE UPDATE ON reports 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Alternative for MySQL (if not using PostgreSQL)
/*
CREATE TABLE reports (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    case_id CHAR(36) NOT NULL,
    examiner_id CHAR(36) NOT NULL,
    scope TEXT,
    objectives TEXT,
    limitations TEXT,
    tools_methods TEXT,
    final_conclusion TEXT,
    evidence_summary TEXT,
    certification_statement TEXT,
    date_examined DATE,
    status ENUM('draft', 'published', 'archived', 'pending', 'reviewed') DEFAULT 'draft',
    version INTEGER NOT NULL DEFAULT 1,
    report_number VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    
    INDEX idx_reports_case_id (case_id),
    INDEX idx_reports_examiner_id (examiner_id),
    INDEX idx_reports_status (status),
    INDEX idx_reports_created_at (created_at),
    INDEX idx_reports_report_number (report_number),
    INDEX idx_reports_date_examined (date_examined)
);
*/