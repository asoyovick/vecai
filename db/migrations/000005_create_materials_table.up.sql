CREATE TABLE suppliers (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  category   TEXT NOT NULL,
  contact    TEXT NOT NULL DEFAULT '',
  phone      TEXT NOT NULL DEFAULT '',
  email      TEXT NOT NULL DEFAULT '',
  rating     REAL NOT NULL DEFAULT 0 CHECK (rating BETWEEN 0 AND 5),
  status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  location   TEXT NOT NULL DEFAULT '',
  last_order TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);

CREATE TABLE materials (
  id            SERIAL PRIMARY KEY,
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
  created_at    TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')),
  updated_at    TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_materials_project ON materials (project_id);
CREATE INDEX idx_materials_status ON materials (status);
