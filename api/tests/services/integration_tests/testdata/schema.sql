-- =========================================================
-- AEGIS - Canonical Test Schema (idempotent, Postgres 16)
-- =========================================================

SET search_path = public;

-- ---------- Extensions ----------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------- Enums (idempotent) ----------
DO $$ 
BEGIN
  CREATE TYPE user_role AS ENUM (
    'Admin','System Admin','DFIR Admin','Tenant Admin','External Collaborator',
    'Incident Responder','Forensic Analyst','Malware Analyst','Threat Intelligent Analyst',
    'DFIR Manager','Detection Engineer','Network Evidence Analyst','Image Forensics Analyst',
    'Disk Forensics Analyst','Log Analyst','Mobile Device Analyst','Memory Forensics Analyst',
    'Cloud Forensics Specialist','Endpoint Forensics Analyst','Reverse Engineer',
    'SIEM Analyst','Vulnerability Analyst','Digital Evidence Technician','Packet Analyst',
    'Legal/Compliance Liaison','Compliance Officer','Legal Counsel','Policy Analyst',
    'SOC Analyst','Incident Commander','Crisis Communications Officer','IT Infrastructure Liaison',
    'Triage Analyst','Evidence Archivist','Training Coordinator','Audit Reviewer','Threat Hunter'
  );
EXCEPTION
  WHEN duplicate_object THEN
    NULL;  -- <-- this semicolon is required
END $$;
-- or: END $$ LANGUAGE plpgsql;


DO $$ BEGIN CREATE TYPE token_status AS ENUM ('active','expired','revoked'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE token_type   AS ENUM ('EMAIL_VERIFY','RESET_PASSWORD','COLLAB_INVITE'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE case_status AS ENUM ('open','under_review','closed','Ongoing','Archived','Under Review','Open'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE case_priority AS ENUM ('low','medium','high','critical','time-sensitive'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE investigation_stage AS ENUM (
  'Triage','Evidence Collection','Analysis','Correlation & Threat Intelligence',
  'Containment & Eradication','Recovery','Reporting & Documentation','Case Closure & Review'
); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE coc_action AS ENUM ('collected','transferred','received','analyzed','checked_in','checked_out','released','disposed'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE report_status AS ENUM ('draft','review','published','archived'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE report_format AS ENUM ('pdf','json','csv'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- Core reference ----------
CREATE TABLE IF NOT EXISTS tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  team_name TEXT NOT NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role user_role,
  is_verified BOOLEAN DEFAULT FALSE,
  profile_picture_url TEXT,
  token_version INTEGER DEFAULT 1,
  external_token_expiry TIMESTAMP,
  external_token_status token_status DEFAULT 'active',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  email_verified_at TIMESTAMP NULL,
  accepted_terms_at TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
  team_id UUID REFERENCES teams(id) ON DELETE SET NULL
);

-- ---------- Auth/Notifications ----------
CREATE TABLE IF NOT EXISTS notifications (
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

CREATE TABLE IF NOT EXISTS tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type token_type NOT NULL,
  token TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP,
  used BOOLEAN DEFAULT FALSE
);

-- ---------- E2EE (X3DH) ----------
CREATE TABLE IF NOT EXISTS x3dh_identity_keys (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  public_key TEXT NOT NULL,
  private_key TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS x3dh_signed_prekeys (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  public_key TEXT NOT NULL,
  private_key TEXT NOT NULL,
  signature TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS x3dh_one_time_prekeys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  public_key TEXT NOT NULL,
  private_key TEXT NOT NULL,
  is_used BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ---------- Cases ----------
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
  team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL
);

-- ---------- Groups / Threads / Messages ----------
CREATE TABLE IF NOT EXISTS groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  case_id UUID REFERENCES cases(id) ON DELETE CASCADE,
  group_url TEXT,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS annotation_threads (
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
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS thread_tags (
  id UUID PRIMARY KEY,
  thread_id UUID NOT NULL REFERENCES annotation_threads(id),
  tag_name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS thread_participants (
  thread_id UUID NOT NULL REFERENCES annotation_threads(id),
  user_id UUID NOT NULL,
  joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (thread_id, user_id)
);

CREATE TABLE IF NOT EXISTS threads (
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

CREATE TABLE IF NOT EXISTS thread_messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  thread_id UUID NOT NULL REFERENCES threads(id),
  parent_message_id UUID,
  user_id UUID NOT NULL REFERENCES users(id),
  message TEXT NOT NULL,
  is_approved BOOLEAN,
  approved_by UUID,
  approved_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
  CONSTRAINT fk_thread_messages_parent FOREIGN KEY (parent_message_id) REFERENCES thread_messages(id)
);

CREATE TABLE IF NOT EXISTS message_mentions (
  message_id UUID NOT NULL REFERENCES thread_messages(id),
  mentioned_user_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
  PRIMARY KEY (message_id, mentioned_user_id)
);

CREATE TABLE IF NOT EXISTS message_reactions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  message_id UUID NOT NULL REFERENCES thread_messages(id),
  user_id UUID NOT NULL REFERENCES users(id),
  reaction TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL
);

-- ---------- Evidence & Tags ----------
CREATE TABLE IF NOT EXISTS evidence (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  uploaded_by UUID NOT NULL REFERENCES users(id),
  filename TEXT NOT NULL,
  file_type TEXT NOT NULL,
  ipfs_cid TEXT NOT NULL,
  file_size INTEGER CHECK (file_size >= 0),
  checksum TEXT NOT NULL,
  metadata JSONB,
  uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  team_id   UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tags (
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS evidence_tags (
  evidence_id UUID REFERENCES evidence(id) ON DELETE CASCADE,
  tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (evidence_id, tag_id)
);

-- ---------- IOC (single, de-duplicated) ----------
CREATE TABLE IF NOT EXISTS iocs (
  id SERIAL PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL,
  value VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT unique_case_ioc UNIQUE (case_id, type, value)
);
CREATE INDEX IF NOT EXISTS idx_iocs_tenant_id ON iocs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iocs_case_id ON iocs(case_id);
CREATE INDEX IF NOT EXISTS idx_iocs_type_value ON iocs(type, value);
CREATE INDEX IF NOT EXISTS idx_iocs_tenant_type_value ON iocs(tenant_id, type, value);

-- ---------- RBAC ----------
CREATE TABLE IF NOT EXISTS permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS enum_role_permissions (
  role user_role NOT NULL,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role user_role NOT NULL,
  PRIMARY KEY (user_id, role)
);

CREATE TABLE IF NOT EXISTS case_user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  role user_role NOT NULL,
  assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
  team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
  PRIMARY KEY (user_id, case_id)
);
CREATE INDEX IF NOT EXISTS idx_case_user_roles_case_id ON case_user_roles(case_id);
CREATE INDEX IF NOT EXISTS idx_case_user_roles_user_id ON case_user_roles(user_id);

-- ---------- Timeline ----------
CREATE TABLE IF NOT EXISTS timeline_events (
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
CREATE INDEX IF NOT EXISTS idx_timeline_case_order ON timeline_events (case_id, "order");

-- ---------- Chain of Custody (fixed & complete) ----------
CREATE TABLE IF NOT EXISTS chain_of_custody (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  evidence_id UUID REFERENCES evidence(id) ON DELETE SET NULL,
  actor_id UUID NOT NULL REFERENCES users(id),
  action coc_action NOT NULL,
  reason TEXT,
  location TEXT,
  hash_md5 TEXT,
  hash_sha1 TEXT,
  hash_sha256 TEXT,
  occurred_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coc_case_time ON chain_of_custody (case_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_coc_evidence_time ON chain_of_custody (evidence_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_coc_actor ON chain_of_custody (actor_id);
CREATE INDEX IF NOT EXISTS idx_coc_action ON chain_of_custody (action);

CREATE OR REPLACE FUNCTION forbid_coc_update_delete()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'Chain of Custody entries are immutable';
END $$;

DROP TRIGGER IF EXISTS trg_coc_no_update ON chain_of_custody;
CREATE TRIGGER trg_coc_no_update
BEFORE UPDATE ON chain_of_custody
FOR EACH ROW EXECUTE FUNCTION forbid_coc_update_delete();

DROP TRIGGER IF EXISTS trg_coc_no_delete ON chain_of_custody;
CREATE TRIGGER trg_coc_no_delete
BEFORE DELETE ON chain_of_custody
FOR EACH ROW EXECUTE FUNCTION forbid_coc_update_delete();

CREATE OR REPLACE VIEW v_chain_of_custody_with_actor AS
SELECT
  c.id,
  c.case_id,
  c.evidence_id,
  c.actor_id,
  u.full_name AS actor_name,
  u.email     AS actor_email,
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

-- ---------- Reports / Case Reports ----------
-- helper for updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END $$;

CREATE TABLE IF NOT EXISTS case_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  examiner_id UUID NOT NULL REFERENCES users(id),
  name VARCHAR(255) NOT NULL,
  file_path VARCHAR(255) NOT NULL,
  scope TEXT,
  objectives TEXT,
  limitations TEXT,
  tools_methods TEXT,
  final_conclusion TEXT,
  evidence_summary TEXT,
  certification_statement TEXT,
  date_examined DATE,
  status report_status NOT NULL DEFAULT 'draft',
  version INTEGER NOT NULL DEFAULT 1,
  report_number TEXT UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_case_reports_case_id ON case_reports(case_id);
CREATE INDEX IF NOT EXISTS idx_case_reports_examiner_id ON case_reports(examiner_id);
CREATE INDEX IF NOT EXISTS idx_case_reports_status ON case_reports(status);

DROP TRIGGER IF EXISTS trg_case_reports_updated_at ON case_reports;
CREATE TRIGGER trg_case_reports_updated_at
BEFORE UPDATE ON case_reports
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- minimal device table so report_hashes FK resolves
CREATE TABLE IF NOT EXISTS report_devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE IF NOT EXISTS report_hashes (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  device_id UUID NOT NULL REFERENCES report_devices(id) ON DELETE CASCADE,
  md5       TEXT,
  sha1      TEXT,
  sha256    TEXT,
  context   TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- main "reports" table
CREATE TABLE IF NOT EXISTS reports (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL,
  team_id   UUID NOT NULL,
  case_id UUID NOT NULL,
  examiner_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  mongo_id CHAR(24),
  report_number VARCHAR(255) NOT NULL,
  status report_status DEFAULT 'draft',
  version INTEGER NOT NULL DEFAULT 1,
  date_examined DATE,
  file_path VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_reports_case_id ON reports(case_id);
CREATE INDEX IF NOT EXISTS idx_reports_examiner_id ON reports(examiner_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);
CREATE INDEX IF NOT EXISTS idx_reports_report_number ON reports(report_number);
CREATE INDEX IF NOT EXISTS idx_reports_date_examined ON reports(date_examined);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS update_reports_updated_at ON reports;
CREATE TRIGGER update_reports_updated_at
BEFORE UPDATE ON reports
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ---------- Evidence indexes ----------
CREATE INDEX IF NOT EXISTS idx_evidence_case_id ON evidence(case_id);
CREATE INDEX IF NOT EXISTS idx_evidence_tenant_id ON evidence(tenant_id);
CREATE INDEX IF NOT EXISTS idx_evidence_team_id ON evidence(team_id);
CREATE INDEX IF NOT EXISTS idx_evidence_checksum ON evidence(checksum);
