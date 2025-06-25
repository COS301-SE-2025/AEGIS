--Aministrative permissions
-- Has all permissions

-- External Collaborator Permissions
INSERT INTO enum_role_permissions (role, permission)
VALUES 
  ('External Collaborator', 'evidence:view'),
  ('External Collaborator', 'case:view'),
  ('External Collaborator', 'comment:create'),
  ('External Collaborator', 'comment:reply'),
  ('External Collaborator', 'chat:send_message'),
  ('External Collaborator', 'chat:view_history'),
  ('External Collaborator', 'thread:create')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('External Collaborator', 'evidence:view'),
  ('External Collaborator', 'case:view'),
  ('External Collaborator', 'comment:create'),
  ('External Collaborator', 'comment:reply'),
  ('External Collaborator', 'chat:send_message'),
  ('External Collaborator', 'chat:view_history'),
  ('External Collaborator', 'thread:create')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Incident Responder', 'case:view'),
--  ('Incident Responder', 'case:update'), reserving for admin
--  ('Incident Responder', 'case:assign_user'),   Admin
  ('Incident Responder', 'evidence:view'),
  ('Incident Responder', 'evidence:verify_checksum'),
  ('Incident Responder', 'comment:create'),
  ('Incident Responder', 'comment:reply'),
  ('Incident Responder', 'chat:send_message'),
  ('Incident Responder', 'chat:view_history'),
  ('Incident Responder', 'thread:create'),
  ('Incident Responder', 'dashboard:view')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Forensic Analyst', 'evidence:view'),
  ('Forensic Analyst', 'evidence:verify_checksum'),
  ('Forensic Analyst', 'evidence:update_metadata'),
  ('Forensic Analyst', 'evidence:tag'),
 -- ('Forensic Analyst', 'evidence:assign_analyst'),  do we need this
  ('Forensic Analyst', 'comment:create'),
  ('Forensic Analyst', 'comment:reply'),
  ('Forensic Analyst', 'thread:create'),
  ('Forensic Analyst', 'dashboard:view'),
  ('Forensic Analyst', 'audit:view_case')
ON CONFLICT DO NOTHING;

-- Malware Analyst
INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Malware Analyst', 'evidence:view'),
  ('Malware Analyst', 'evidence:verify_checksum'),
  ('Malware Analyst', 'evidence:tag'),
  ('Malware Analyst', 'comment:create'),
  ('Malware Analyst', 'comment:reply'),
  ('Malware Analyst', 'thread:create'),
  ('Malware Analyst', 'dashboard:view'),
  ('Malware Analyst', 'audit:view_case')
ON CONFLICT DO NOTHING;

-- Threat Intelligent Analyst
INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Threat Intelligent Analyst', 'evidence:view'),
  ('Threat Intelligent Analyst', 'evidence:tag'),
  ('Threat Intelligent Analyst', 'case:view'),
  ('Threat Intelligent Analyst', 'comment:create'),
  ('Threat Intelligent Analyst', 'comment:reply'),
  ('Threat Intelligent Analyst', 'thread:create'),
  ('Threat Intelligent Analyst', 'dashboard:view'),
  ('Threat Intelligent Analyst', 'audit:view_case')
ON CONFLICT DO NOTHING;


-- DFIR Manager
INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('DFIR Manager', 'case:view'),
  ('DFIR Manager', 'case:update'),
  ('DFIR Manager', 'case:assign_user'),
  ('DFIR Manager', 'case:change_status'),
  ('DFIR Manager', 'case:set_priority'),
  ('DFIR Manager', 'evidence:view'),
  ('DFIR Manager', 'evidence:assign_analyst'),
  ('DFIR Manager', 'dashboard:view'),
  ('DFIR Manager', 'audit:view_case'),
  ('DFIR Manager', 'chat:create_room'),
  ('DFIR Manager', 'chat:moderate')
ON CONFLICT DO NOTHING;

-- Detection Engineer
INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Detection Engineer', 'evidence:view'),
  ('Detection Engineer', 'evidence:tag'),
  ('Detection Engineer', 'case:view'),
  ('Detection Engineer', 'comment:create'),
  ('Detection Engineer', 'comment:reply'),
  ('Detection Engineer', 'thread:create'),
  ('Detection Engineer', 'dashboard:view'),
  ('Detection Engineer', 'audit:view_case')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Network Evidence Analyst', 'evidence:view'),
  ('Network Evidence Analyst', 'evidence:verify_checksum'),
  ('Network Evidence Analyst', 'evidence:tag'),
  ('Network Evidence Analyst', 'comment:create'),
  ('Network Evidence Analyst', 'comment:reply'),
  ('Network Evidence Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Image Forensics Analyst', 'evidence:view'),
  ('Image Forensics Analyst', 'evidence:verify_checksum'),
  ('Image Forensics Analyst', 'evidence:tag'),
  ('Image Forensics Analyst', 'comment:create'),
  ('Image Forensics Analyst', 'comment:reply'),
  ('Image Forensics Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Disk Forensics Analyst', 'evidence:view'),
  ('Disk Forensics Analyst', 'evidence:verify_checksum'),
  ('Disk Forensics Analyst', 'evidence:tag'),
  ('Disk Forensics Analyst', 'comment:create'),
  ('Disk Forensics Analyst', 'comment:reply'),
  ('Disk Forensics Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Log Analyst', 'evidence:view'),
  ('Log Analyst', 'evidence:verify_checksum'),
  ('Log Analyst', 'evidence:tag'),
  ('Log Analyst', 'audit:view_all'),
  --('Log Analyst', 'audit:view_case'),
 -- ('Log Analyst', 'audit:view_user'),
  ('Log Analyst', 'comment:create'),
  ('Log Analyst', 'comment:reply'),
  ('Log Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Mobile Device Analyst', 'evidence:view'),
  ('Mobile Device Analyst', 'evidence:verify_checksum'),
  ('Mobile Device Analyst', 'evidence:tag'),
  ('Mobile Device Analyst', 'comment:create'),
  ('Mobile Device Analyst', 'comment:reply'),
  ('Mobile Device Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Memory Forensics Analyst', 'evidence:view'),
  ('Memory Forensics Analyst', 'evidence:verify_checksum'),
  ('Memory Forensics Analyst', 'evidence:tag'),
  ('Memory Forensics Analyst', 'audit:view_case'),
  ('Memory Forensics Analyst', 'comment:create'),
  ('Memory Forensics Analyst', 'comment:reply'),
  ('Memory Forensics Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Cloud Forensics Specialist', 'evidence:view'),
  ('Cloud Forensics Specialist', 'evidence:verify_checksum'),
  ('Cloud Forensics Specialist', 'evidence:tag'),
  ('Cloud Forensics Specialist', 'audit:view_case'),
  ('Cloud Forensics Specialist', 'case:view'),
  ('Cloud Forensics Specialist', 'comment:create'),
  ('Cloud Forensics Specialist', 'comment:reply'),
  ('Cloud Forensics Specialist', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Endpoint Forensics Analyst', 'evidence:view'),
  ('Endpoint Forensics Analyst', 'evidence:verify_checksum'),
  ('Endpoint Forensics Analyst', 'evidence:tag'),
  ('Endpoint Forensics Analyst', 'comment:create'),
  ('Endpoint Forensics Analyst', 'comment:reply'),
  ('Endpoint Forensics Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Reverse Engineer', 'evidence:view'),
  ('Reverse Engineer', 'evidence:verify_checksum'),
  ('Reverse Engineer', 'evidence:tag'),
  ('Reverse Engineer', 'comment:create'),
  ('Reverse Engineer', 'comment:reply'),
  ('Reverse Engineer', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('SIEM Analyst', 'dashboard:view'),
  ('SIEM Analyst', 'case:view'),
  ('SIEM Analyst', 'evidence:view'),
  ('SIEM Analyst', 'audit:view_case'),
  ('SIEM Analyst', 'comment:create'),
  ('SIEM Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Vulnerability Analyst', 'case:view'),
  ('Vulnerability Analyst', 'evidence:view'),
  ('Vulnerability Analyst', 'evidence:verify_checksum'),
  ('Vulnerability Analyst', 'comment:create'),
  ('Vulnerability Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Digital Evidence Technician', 'evidence:view'),
  ('Digital Evidence Technician', 'evidence:verify_checksum'),
  ('Digital Evidence Technician', 'evidence:upload'),
  ('Digital Evidence Technician', 'evidence:tag'),
  ('Digital Evidence Technician', 'comment:create'),
  ('Digital Evidence Technician', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Packet Analyst', 'evidence:view'),
  ('Packet Analyst', 'evidence:verify_checksum'),
  ('Packet Analyst', 'evidence:tag'),
  ('Packet Analyst', 'comment:create'),
  ('Packet Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Legal/Compliance Liaison', 'case:view'),
  ('Legal/Compliance Liaison', 'evidence:view'),
  ('Legal/Compliance Liaison', 'audit:view_case'),
  ('Legal/Compliance Liaison', 'comment:create'),
  ('Legal/Compliance Liaison', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Compliance Officer', 'case:view'),
  ('Compliance Officer', 'audit:view_all'),
  ('Compliance Officer', 'comment:create'),
  ('Compliance Officer', 'chat:send_message')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Legal Counsel', 'case:view'),
  ('Legal Counsel', 'evidence:view'),
  ('Legal Counsel', 'comment:create'),
  ('Legal Counsel', 'chat:send_message')
ON CONFLICT DO NOTHING;


INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Policy Analyst', 'case:view'),
  ('Policy Analyst', 'dashboard:view'),
  ('Policy Analyst', 'comment:create'),
  ('Policy Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('SOC Analyst', 'evidence:view'),
  ('SOC Analyst', 'evidence:verify_checksum'),
  ('SOC Analyst', 'case:view'),
  ('SOC Analyst', 'comment:create'),
  ('SOC Analyst', 'chat:send_message'),
  ('SOC Analyst', 'dashboard:view')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Incident Commander', 'case:view'),
  ('Incident Commander', 'case:change_status'),
  ('Incident Commander', 'case:set_priority'),
  ('Incident Commander', 'audit:view_case'),
  ('Incident Commander', 'comment:create'),
  ('Incident Commander', 'chat:create_room'),
  ('Incident Commander', 'chat:send_message'),
  ('Incident Commander', 'dashboard:view')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Crisis Communications Officer', 'case:view'),
  ('Crisis Communications Officer', 'comment:create'),
  ('Crisis Communications Officer', 'comment:reply'),
  ('Crisis Communications Officer', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('IT Infrastructure Liaison', 'case:view'),
  ('IT Infrastructure Liaison', 'evidence:view'),
  ('IT Infrastructure Liaison', 'comment:create'),
  ('IT Infrastructure Liaison', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Triage Analyst', 'case:view'),
  ('Triage Analyst', 'evidence:view'),
  ('Triage Analyst', 'evidence:verify_checksum'),
  ('Triage Analyst', 'comment:create'),
  ('Triage Analyst', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Evidence Archivist', 'evidence:view'),
  ('Evidence Archivist', 'evidence:verify_checksum'),
  ('Evidence Archivist', 'dashboard:view'),
  ('Evidence Archivist', 'comment:create'),
  ('Evidence Archivist', 'comment:reply')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Training Coordinator', 'case:view'),
  ('Training Coordinator', 'comment:create'),
  ('Training Coordinator', 'chat:send_message')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Audit Reviewer', 'audit:view_all'),
  ('Audit Reviewer', 'audit:view_case'),
  ('Audit Reviewer', 'audit:view_user'),
  ('Audit Reviewer', 'audit:generate_report'),
  ('Audit Reviewer', 'dashboard:view')
ON CONFLICT DO NOTHING;

INSERT INTO enum_role_permissions (role, permission)
VALUES
  ('Threat Hunter', 'evidence:view'),
  ('Threat Hunter', 'evidence:verify_checksum'),
  ('Threat Hunter', 'case:view'),
  ('Threat Hunter', 'dashboard:view'),
  ('Threat Hunter', 'comment:create'),
  ('Threat Hunter', 'chat:send_message')
ON CONFLICT DO NOTHING;