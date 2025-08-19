-- === AEGIS Reporting Module: Finalized Schema (validated) ===
--****************************
-- 1.  docker cp ./reporting.sql aegis-postgres-1:/reporting.sql

-- 2. docker exec -it aegis-postgres-1 psql -U app_user -d app_database -f /reporting.sql

-- Prereq: UUID generation (gen_random_uuid)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------- Enums ----------
DO $$ BEGIN
  CREATE TYPE report_status  AS ENUM ('draft', 'published', 'review', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE report_format  AS ENUM ('pdf','json','csv');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE finding_category AS ENUM (
    'email','chat','browser','filesystem','timeline',
    'deleted_recovery','ioc','encryption','other'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE severity_level AS ENUM ('info','low','medium','high','critical');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE appendix_type  AS ENUM (
    'screenshot','extracted_data','log','hash_listing',
    'custody_document','glossary','cv','other'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- Common trigger ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END $$;

-- =====================================================================
-- CORE REPORT RECORD (prose + lifecycle)
-- =====================================================================
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

CREATE TRIGGER trg_case_reports_updated_at
BEFORE UPDATE ON case_reports
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- At most one FINAL per case
CREATE UNIQUE INDEX IF NOT EXISTS uq_case_reports_one_final_per_case
ON case_reports(case_id) WHERE status = 'final';

-- =====================================================================
-- EVIDENCE SUMMARY DETAIL (devices + integrity hashes)
-- =====================================================================
CREATE TABLE IF NOT EXISTS report_devices (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id      UUID NOT NULL REFERENCES case_reports(id) ON DELETE CASCADE,
  evidence_id    UUID REFERENCES evidence(id) ON DELETE SET NULL,
  device_type    TEXT,
  make           TEXT,
  model          TEXT,
  serial_number  TEXT,
  description    TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_report_devices_report_id ON report_devices(report_id);

CREATE TRIGGER trg_report_devices_updated_at
BEFORE UPDATE ON report_devices
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS report_hashes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  device_id   UUID NOT NULL REFERENCES report_devices(id) ON DELETE CASCADE,
  md5         TEXT,
  sha1        TEXT,
  sha256      TEXT,
  context     TEXT,  -- "acquisition image", "working copy", etc.
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_report_hashes_device_id ON report_hashes(device_id);

-- =====================================================================
-- FINDINGS (optionally linked to evidence and report)
-- =====================================================================
CREATE TABLE IF NOT EXISTS findings (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id        UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  report_id      UUID REFERENCES case_reports(id) ON DELETE SET NULL,
  evidence_id    UUID REFERENCES evidence(id) ON DELETE SET NULL,
  category       finding_category NOT NULL,
  title          TEXT NOT NULL,
  detail         TEXT NOT NULL,
  severity       severity_level DEFAULT 'info',
  related_user   TEXT,
  related_path   TEXT,
  ts_range       TSTZRANGE,    -- time interval, if applicable
  created_by     UUID REFERENCES users(id),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_findings_case_id      ON findings(case_id);
CREATE INDEX IF NOT EXISTS idx_findings_report_id    ON findings(report_id);
CREATE INDEX IF NOT EXISTS idx_findings_evidence_id  ON findings(evidence_id);
CREATE INDEX IF NOT EXISTS idx_findings_ts_range     ON findings USING GIST (ts_range);

CREATE TRIGGER trg_findings_updated_at
BEFORE UPDATE ON findings
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================================
-- APPENDICES & EVIDENCE SCREENSHOTS
-- =====================================================================
CREATE TABLE IF NOT EXISTS appendices (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id       UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  report_id     UUID REFERENCES case_reports(id) ON DELETE SET NULL,
  appendix_type appendix_type NOT NULL,
  title         TEXT,
  description   TEXT,
  file_ref      TEXT NOT NULL,  -- object-store key / path / URL
  created_by    UUID REFERENCES users(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_appendices_case_id   ON appendices(case_id);
CREATE INDEX IF NOT EXISTS idx_appendices_report_id ON appendices(report_id);

CREATE TABLE IF NOT EXISTS evidence_screenshots (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id       UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  evidence_id   UUID REFERENCES evidence(id) ON DELETE SET NULL,
  caption       TEXT,
  source        TEXT,         -- "FTK", "Autopsy", etc.
  file_ref      TEXT NOT NULL,
  taken_at      TIMESTAMPTZ,
  created_by    UUID REFERENCES users(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shots_case_id     ON evidence_screenshots(case_id);
CREATE INDEX IF NOT EXISTS idx_shots_evidence_id ON evidence_screenshots(evidence_id);

-- =====================================================================
-- GENERATED REPORT ARTIFACTS (PDF/JSON/CSV + SHA-256)
-- =====================================================================
CREATE TABLE IF NOT EXISTS report_artifacts (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id       UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  report_id     UUID NOT NULL REFERENCES case_reports(id) ON DELETE CASCADE,
  format        report_format NOT NULL,
  storage_ref   TEXT NOT NULL,    -- object key / URL / path
  size_bytes    BIGINT,
  sha256        TEXT,             -- integrity
  generated_by  UUID NOT NULL REFERENCES users(id),
  generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_report_artifacts_case_id   ON report_artifacts(case_id);
CREATE INDEX IF NOT EXISTS idx_report_artifacts_report_id ON report_artifacts(report_id);
CREATE INDEX IF NOT EXISTS idx_report_artifacts_format    ON report_artifacts(format);

-- =====================================================================
-- CHAIN OF CUSTODY (append-only)
-- =====================================================================
CREATE TABLE IF NOT EXISTS chain_of_custody (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id      UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  evidence_id  UUID NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
  actor_id     UUID REFERENCES users(id),
  action       TEXT NOT NULL,      -- "upload","download","analysis","transfer","seal",...
  reason       TEXT,
  location     TEXT,
  hash_md5     TEXT,
  hash_sha1    TEXT,
  hash_sha256  TEXT,
  occurred_at  TIMESTAMPTZ NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coc_case_id      ON chain_of_custody(case_id);
CREATE INDEX IF NOT EXISTS idx_coc_evidence_id  ON chain_of_custody(evidence_id);
CREATE INDEX IF NOT EXISTS idx_coc_actor_id     ON chain_of_custody(actor_id);
CREATE INDEX IF NOT EXISTS idx_coc_occurred_at  ON chain_of_custody(occurred_at);

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
