CREATE TABLE project_documents (
  id            SERIAL PRIMARY KEY,
  project_id    INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  uploaded_by   INTEGER REFERENCES users (id) ON DELETE SET NULL,
  name          TEXT    NOT NULL,
  category      TEXT    NOT NULL
                CHECK (category IN ('Architectural Drawing','Structural Drawing','BOQ','Contract','Permit','Invoice','Site Report','Inspection Report','Photo')),
  version       TEXT    NOT NULL DEFAULT '1.0',
  original_name TEXT    NOT NULL,
  stored_name   TEXT    NOT NULL,
  mime_type     TEXT    NOT NULL DEFAULT 'application/octet-stream',
  size_bytes    INTEGER NOT NULL DEFAULT 0,
  format        TEXT    NOT NULL DEFAULT '',
  created_at    TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_project_documents_project ON project_documents (project_id);
CREATE INDEX idx_project_documents_category ON project_documents (category);
