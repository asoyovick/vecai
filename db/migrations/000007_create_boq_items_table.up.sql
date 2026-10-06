CREATE TABLE boq_items (
  id          SERIAL PRIMARY KEY,
  project_id  INTEGER NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  item_no     TEXT    NOT NULL DEFAULT '',
  description TEXT    NOT NULL,
  unit        TEXT    NOT NULL DEFAULT 'pcs',
  quantity    REAL    NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  unit_rate   INTEGER NOT NULL DEFAULT 0 CHECK (unit_rate >= 0),
  created_at  TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')),
  updated_at  TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_boq_items_project ON boq_items (project_id);
