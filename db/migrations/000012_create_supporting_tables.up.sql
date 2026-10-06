CREATE TABLE sessions (
  token_hash TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')),
  expires_at TEXT NOT NULL
);
CREATE INDEX idx_sessions_user ON sessions (user_id);

CREATE TABLE organisation (
  id            INTEGER PRIMARY KEY CHECK (id = 1),
  name          TEXT NOT NULL,
  business_type TEXT NOT NULL DEFAULT 'Construction Company',
  country       TEXT NOT NULL DEFAULT 'Kenya',
  currency      TEXT NOT NULL DEFAULT 'KES — Kenyan Shilling'
);

CREATE TABLE material_deliveries (
  id          SERIAL PRIMARY KEY,
  material_id INTEGER NOT NULL REFERENCES materials (id) ON DELETE CASCADE,
  quantity    REAL    NOT NULL CHECK (quantity > 0),
  note        TEXT    NOT NULL DEFAULT '',
  recorded_by INTEGER REFERENCES users (id) ON DELETE SET NULL,
  created_at  TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_deliveries_material ON material_deliveries (material_id);

CREATE TABLE tasks (
  id          SERIAL PRIMARY KEY,
  project_id  INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  assignee_id INTEGER REFERENCES users (id) ON DELETE SET NULL,
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  start_date  TEXT,
  due_date    TEXT,
  status      TEXT    NOT NULL DEFAULT 'not-started' CHECK (status IN ('not-started','in-progress','blocked','completed')),
  priority    TEXT    NOT NULL DEFAULT 'medium' CHECK (priority IN ('high','medium','low')),
  created_at  TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')),
  updated_at  TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_tasks_project ON tasks (project_id);
CREATE INDEX idx_tasks_status ON tasks (status);
CREATE INDEX idx_tasks_assignee ON tasks (assignee_id);

CREATE TABLE daily_report_materials (
  id             SERIAL PRIMARY KEY,
  daily_report_id INTEGER NOT NULL REFERENCES daily_reports (id) ON DELETE CASCADE,
  name           TEXT NOT NULL,
  quantity       REAL NOT NULL DEFAULT 0,
  unit           TEXT NOT NULL DEFAULT ''
);

CREATE TABLE budget_months (
  project_id INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  ym         TEXT    NOT NULL,
  planned    INTEGER NOT NULL DEFAULT 0,
  actual     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (project_id, ym)
);

CREATE TABLE activity_log (
  id         SERIAL PRIMARY KEY,
  project_id INTEGER REFERENCES projects (id) ON DELETE CASCADE,
  actor_id   INTEGER REFERENCES users (id) ON DELETE SET NULL,
  actor_name TEXT NOT NULL DEFAULT 'System',
  type       TEXT NOT NULL CHECK (type IN ('task','document','material','site-update','project','comment','system')),
  text       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_activity_project ON activity_log (project_id, created_at);
CREATE INDEX idx_activity_created ON activity_log (created_at);

CREATE TABLE user_preferences (
  user_id INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  prefs   TEXT NOT NULL DEFAULT '{}'
);
