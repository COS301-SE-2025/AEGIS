-- =============================
-- RBAC + Case User Roles Schema (ENUM-based, simplified)
-- =============================


--****************************
-- 1.  docker cp ./schema.sql aegis-postgres-1:/schema.sql

-- 2. docker exec -it aegis-postgres-1 psql -U app_user -d app_database -f /schema.sql

--****************************
-- Create ENUM types
CREATE TYPE user_role AS ENUM (
    'Incident Responder', 
    'Forensic Analyst', 
    'Malware Analyst', 
    'Threat Intelligent Analyst', 
    'DFIR Manager', 
    'Legal/Compliance Liaison', 
    'Detection Engineer', 
    'Generic user'
);

CREATE TYPE case_status AS ENUM ('open', 'under_review', 'closed');
CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE investigation_stage AS ENUM ('analysis', 'research', 'evaluation', 'finalization');

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role,
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token TEXT,
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
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
    metadata JSONB,
    uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

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
--*****************************
--DEMO 2: RBAC with ENUMs for user roles
--****************************
------------------------
--****************************
-- Permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL
);

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
    PRIMARY KEY (user_id, case_id)
);

-- Indexes to support performance (optional but recommended)
CREATE INDEX idx_case_user_roles_case_id ON case_user_roles(case_id);
CREATE INDEX idx_case_user_roles_user_id ON case_user_roles(user_id);
