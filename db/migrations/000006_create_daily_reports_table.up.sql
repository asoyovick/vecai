CREATE TABLE daily_reports (
  id              SERIAL PRIMARY KEY,
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
  created_at      TEXT    NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
);
CREATE INDEX idx_daily_reports_project ON daily_reports (project_id, date);
