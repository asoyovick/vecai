-- VECAI schema — initial migration
CREATE TABLE users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  email         TEXT    NOT NULL UNIQUE COLLATE NOCASE,
  password_hash TEXT    NOT NULL,
  name          TEXT    NOT NULL,
  role          TEXT    NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member', 'viewer')),
  job_title     TEXT    NOT NULL DEFAULT '',
  initials      TEXT    NOT NULL DEFAULT '',
  color         TEXT    NOT NULL DEFAULT '#1e3a5f',
  phone         TEXT    NOT NULL DEFAULT '',
  created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE sessions (
  token_hash TEXT    PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at TEXT    NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT    NOT NULL
);
CREATE INDEX idx_sessions_user ON sessions (user_id);

CREATE TABLE organisation (
  id            INTEGER PRIMARY KEY CHECK (id = 1),
  name          TEXT NOT NULL,
  business_type TEXT NOT NULL DEFAULT 'Construction Company',
  country       TEXT NOT NULL DEFAULT 'Kenya',
  currency      TEXT NOT NULL DEFAULT 'KES — Kenyan Shilling'
);

CREATE TABLE projects (
  id                  INTEGER PRIMARY KEY AUTOINCREMENT,
  name                TEXT    NOT NULL,
  location            TEXT    NOT NULL DEFAULT '',
  status              TEXT    NOT NULL DEFAULT 'Tendering'
                      CHECK (status IN ('Tendering','Design','Foundation','Construction','Roofing','Finishing','Handover','Completed')),
  progress            INTEGER NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
  start_date          TEXT,
  expected_completion TEXT,
  client              TEXT    NOT NULL DEFAULT '',
  contractor          TEXT    NOT NULL DEFAULT '',
  budget_total        INTEGER NOT NULL DEFAULT 0 CHECK (budget_total >= 0),
  budget_spent        INTEGER NOT NULL DEFAULT 0 CHECK (budget_spent >= 0),
  description         TEXT    NOT NULL DEFAULT '',
  image_url           TEXT    NOT NULL DEFAULT '',
  created_at          TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at          TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_projects_status ON projects (status);

CREATE TABLE project_members (
  project_id INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, user_id)
);
CREATE INDEX idx_project_members_user ON project_members (user_id);

CREATE TABLE suppliers (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL,
  category   TEXT NOT NULL,
  contact    TEXT NOT NULL DEFAULT '',
  phone      TEXT NOT NULL DEFAULT '',
  email      TEXT NOT NULL DEFAULT '',
  rating     REAL NOT NULL DEFAULT 0 CHECK (rating BETWEEN 0 AND 5),
  status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  location   TEXT NOT NULL DEFAULT '',
  last_order TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE materials (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id    INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  supplier_id   INTEGER REFERENCES suppliers (id) ON DELETE SET NULL,
  name          TEXT    NOT NULL,
  unit          TEXT    NOT NULL DEFAULT 'pcs',
  required_qty  REAL    NOT NULL DEFAULT 0 CHECK (required_qty >= 0),
  available_qty REAL    NOT NULL DEFAULT 0 CHECK (available_qty >= 0),
  used_qty      REAL    NOT NULL DEFAULT 0 CHECK (used_qty >= 0),
  status        TEXT    NOT NULL DEFAULT 'ok' CHECK (status IN ('ok', 'low', 'critical', 'ordered')),
  unit_cost     INTEGER NOT NULL DEFAULT 0 CHECK (unit_cost >= 0),
  last_delivery TEXT,
  created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_materials_project ON materials (project_id);
CREATE INDEX idx_materials_status ON materials (status);

CREATE TABLE material_deliveries (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  material_id INTEGER NOT NULL REFERENCES materials (id) ON DELETE CASCADE,
  quantity    REAL    NOT NULL CHECK (quantity > 0),
  note        TEXT    NOT NULL DEFAULT '',
  recorded_by INTEGER REFERENCES users (id) ON DELETE SET NULL,
  created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_deliveries_material ON material_deliveries (material_id);

CREATE TABLE tasks (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id  INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  assignee_id INTEGER REFERENCES users (id) ON DELETE SET NULL,
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  start_date  TEXT,
  due_date    TEXT,
  status      TEXT    NOT NULL DEFAULT 'not-started' CHECK (status IN ('not-started','in-progress','blocked','completed')),
  priority    TEXT    NOT NULL DEFAULT 'medium' CHECK (priority IN ('high','medium','low')),
  created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_tasks_project ON tasks (project_id);
CREATE INDEX idx_tasks_status ON tasks (status);
CREATE INDEX idx_tasks_assignee ON tasks (assignee_id);

CREATE TABLE documents (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
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
  created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_documents_project ON documents (project_id);
CREATE INDEX idx_documents_category ON documents (category);

CREATE TABLE site_updates (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id      INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  submitted_by    INTEGER REFERENCES users (id) ON DELETE SET NULL,
  date            TEXT    NOT NULL,
  weather         TEXT    NOT NULL DEFAULT '',
  temperature     TEXT    NOT NULL DEFAULT '',
  workers_on_site INTEGER NOT NULL DEFAULT 0 CHECK (workers_on_site >= 0),
  work_completed  TEXT    NOT NULL,
  issues          TEXT    NOT NULL DEFAULT '',
  notes           TEXT    NOT NULL DEFAULT '',
  photo_count     INTEGER NOT NULL DEFAULT 0,
  created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_site_updates_project ON site_updates (project_id, date);

CREATE TABLE site_update_materials (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  site_update_id INTEGER NOT NULL REFERENCES site_updates (id) ON DELETE CASCADE,
  name           TEXT NOT NULL,
  quantity       REAL NOT NULL DEFAULT 0,
  unit           TEXT NOT NULL DEFAULT ''
);

CREATE TABLE budget_months (
  project_id INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  ym         TEXT    NOT NULL, -- 'YYYY-MM'
  planned    INTEGER NOT NULL DEFAULT 0,
  actual     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (project_id, ym)
);

CREATE TABLE activity_log (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id INTEGER REFERENCES projects (id) ON DELETE CASCADE,
  actor_id   INTEGER REFERENCES users (id) ON DELETE SET NULL,
  actor_name TEXT NOT NULL DEFAULT 'System',
  type       TEXT NOT NULL CHECK (type IN ('task','document','material','site-update','project','comment','system')),
  text       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_activity_project ON activity_log (project_id, created_at);
CREATE INDEX idx_activity_created ON activity_log (created_at);

CREATE TABLE user_preferences (
  user_id INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  prefs   TEXT NOT NULL DEFAULT '{}'
);
