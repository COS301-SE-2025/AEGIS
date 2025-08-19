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

CREATE TYPE case_status AS ENUM ('open','Open', 'under_review','Under Review', 'closed','Ongoing','Archived');
CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'critical','time-sensitive');
CREATE TYPE investigation_stage_new AS ENUM (
    'Triage',
    'Evidence Collection',
    'Analysis',
    'Correlation & Threat Intelligence',
    'Containment & Eradication',
    'Recovery',
    'Reporting & Documentation',
    'Case Closure & Review'
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
    investigation_stage investigation_stage DEFAULT 'analysis',
    priority case_priority DEFAULT 'medium',
    team_name TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- New Fields
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
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL, -- Link to tenant
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
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    title varchar(255) NOT NULL,
    file_id uuid NOT NULL,
    case_id uuid NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status varchar(50) DEFAULT 'open',
    priority varchar(50) DEFAULT 'medium',
    is_active boolean DEFAULT true,
    resolved_at timestamp DEFAULT NULL
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
    uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    -- New Fields
    -- 🔹 Multi-tenancy fields
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    team_id   UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE
);

--IOCS
CREATE TABLE iocs (
    id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    case_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_case FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,

    -- Prevent duplicate IOC entries per case
    CONSTRAINT unique_case_ioc UNIQUE (case_id, type, value)
);

-- Indexes for performance
CREATE INDEX idx_iocs_tenant_id ON iocs(tenant_id);
CREATE INDEX idx_iocs_case_id ON iocs(case_id);
CREATE INDEX idx_iocs_type_value ON iocs(type, value);
CREATE INDEX idx_iocs_tenant_type_value ON iocs(tenant_id, type, value);


-- Optional indexes for performance
CREATE INDEX idx_evidence_metadata_evidence_id ON evidence_metadata(evidence_id);
CREATE INDEX idx_evidence_metadata_key ON evidence_metadata(key);

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
--****************************
-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL
);

-- Insert: User Management
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'user:create', 'Create/register a new user'),
(gen_random_uuid(), 'user:view', 'View user details'),
(gen_random_uuid(), 'user:update', 'Edit user details'),
(gen_random_uuid(), 'user:delete', 'Remove a user from the system'),
(gen_random_uuid(), 'user:assign_role', 'Assign roles to users'),
(gen_random_uuid(), 'user:verify', 'Manually verify a user''s account'),
(gen_random_uuid(), 'user:reset_password', 'Initiate password reset for a user');

-- Insert: Evidence Management
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'evidence:upload', 'Upload new evidence file'),
(gen_random_uuid(), 'evidence:view', 'View/download evidence'),
(gen_random_uuid(), 'evidence:delete', 'Delete evidence (soft/hard delete)'),
(gen_random_uuid(), 'evidence:update_metadata', 'Edit metadata fields for evidence'),
(gen_random_uuid(), 'evidence:tag', 'Assign tags to evidence'),
(gen_random_uuid(), 'evidence:assign_analyst', 'Assign an analyst to a specific evidence item'),
(gen_random_uuid(), 'evidence:verify_checksum', 'Validate integrity of uploaded file'),
(gen_random_uuid(), 'evidence:view_logs', 'View audit logs related to evidence actions');

-- Extend Evidence Tagging
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'evidence:create_tag', 'Create new tags for evidence'),
(gen_random_uuid(), 'evidence:remove_tag', 'Remove existing tags from evidence');

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
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'case:create', 'Create a new case'),
(gen_random_uuid(), 'case:view', 'View case details'),
(gen_random_uuid(), 'case:update', 'Update case information'),
(gen_random_uuid(), 'case:assign_user', 'Assign user to case'),
(gen_random_uuid(), 'case:change_status', 'Change case status (open, review, closed)'),
(gen_random_uuid(), 'case:set_priority', 'Set or modify case priority'),
(gen_random_uuid(), 'case:archive', 'Archive or lock case for compliance purposes');

-- Case Tagging
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'case:tag', 'Assign tags to cases'),
(gen_random_uuid(), 'case:create_tag', 'Create new tags for cases'),
(gen_random_uuid(), 'case:remove_tag', 'Remove existing tags from cases');


-- Insert: Dashboard & Audit
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'dashboard:view', 'View dashboards with case/evidence summaries'),
(gen_random_uuid(), 'audit:view_all', 'View all system-level audit logs'),
(gen_random_uuid(), 'audit:view_case', 'View audit logs related to a specific case'),
(gen_random_uuid(), 'audit:view_user', 'View audit actions performed by a user'),
(gen_random_uuid(), 'audit:generate_report', 'Export/print audit or investigation report');

-- Insert: Compliance & Legal
-- INSERT INTO permissions (id, name, description) VALUES
-- (gen_random_uuid(), 'compliance:review', 'View compliance indicators for cases'),
-- (gen_random_uuid(), 'compliance:lock_evidence', 'Lock evidence to preserve legal integrity'),
-- (gen_random_uuid(), 'compliance:export_data', 'Export data for legal or regulatory use'),
-- (gen_random_uuid(), 'legal:comment', 'Add legal notes or directives to case/evidence'),
-- (gen_random_uuid(), 'legal:mark_sensitive', 'Mark evidence or case as sensitive/privileged');


-- Log Exporting
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'audit:export_case_log', 'Export audit logs for a specific case'),
(gen_random_uuid(), 'audit:export_user_log', 'Export audit logs for a specific user');

-- Annotation and Commenting
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'comment:create', 'Add a comment to a case or evidence'),
(gen_random_uuid(), 'comment:reply', 'Reply to an existing comment thread'),
(gen_random_uuid(), 'comment:edit', 'Edit your own comment'),
(gen_random_uuid(), 'comment:delete', 'Delete a comment you authored'),
(gen_random_uuid(), 'comment:moderate', 'Moderate or remove inappropriate comments'),
(gen_random_uuid(), 'thread:create', 'Start a new annotation thread on a case or file'),
(gen_random_uuid(), 'thread:close', 'Close or archive a thread to prevent new replies');

-- Case Collaboration
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'collaboration:add_member', 'Add member to a case collaboration team'),
(gen_random_uuid(), 'collaboration:remove_member', 'Remove member from a case team'),
(gen_random_uuid(), 'collaboration:change_role', 'Update user role within a case'),
(gen_random_uuid(), 'collaboration:message', 'Send secure messages within a case context');

-- Secure Chat
INSERT INTO permissions (id, name, description) VALUES
(gen_random_uuid(), 'chat:create_room', 'Create a new chat room for a case'),
(gen_random_uuid(), 'chat:send_message', 'Send a message in a secure chat'),
(gen_random_uuid(), 'chat:view_history', 'View chat history'),
(gen_random_uuid(), 'chat:delete_message', 'Delete your message from chat'),
(gen_random_uuid(), 'chat:moderate', 'Moderate chat content or users');

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

-- iocs related tables
CREATE TABLE iocs (
    id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    case_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_case FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,

    -- Prevent duplicate IOC entries per case
    CONSTRAINT unique_case_ioc UNIQUE (case_id, type, value)
);

-- Indexes for performance
CREATE INDEX idx_iocs_tenant_id ON iocs(tenant_id);
CREATE INDEX idx_iocs_case_id ON iocs(case_id);
CREATE INDEX idx_iocs_type_value ON iocs(type, value);
CREATE INDEX idx_iocs_tenant_type_value ON iocs(tenant_id, type, value);


--Timeline Events Table
CREATE TABLE timeline_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id uuid NOT NULL REFERENCES cases(id),
  description text NOT NULL,
  evidence jsonb NOT NULL DEFAULT '[]'::jsonb,
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  severity varchar(20),
  analyst_id uuid,
  analyst_name varchar(255),
  "order" integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX idx_timeline_case_order ON timeline_events (case_id, "order");

----Chain of Custody Entries table-----
CREATE TABLE chain_of_custody (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evidence_id UUID NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    custodian TEXT NOT NULL,
    acquisition_date TIMESTAMP WITH TIME ZONE,
    acquisition_tool TEXT,
    system_info JSONB,  -- os_version, architecture, computer_name, domain, etc.
    forensic_info JSONB, -- method, examiner, location, notes, legal_status
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

