ALTER TABLE projects ADD COLUMN status TEXT NOT NULL DEFAULT 'Tendering'
  CHECK (status IN ('Tendering','Design','Foundation','Construction','Roofing','Finishing','Handover','Completed'));
ALTER TABLE projects ADD COLUMN progress INTEGER NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100);
ALTER TABLE projects ADD COLUMN start_date TEXT;
ALTER TABLE projects ADD COLUMN expected_completion TEXT;
ALTER TABLE projects ADD COLUMN client TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN contractor TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN budget_total INTEGER NOT NULL DEFAULT 0 CHECK (budget_total >= 0);
ALTER TABLE projects ADD COLUMN budget_spent INTEGER NOT NULL DEFAULT 0 CHECK (budget_spent >= 0);
ALTER TABLE projects ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN image_url TEXT NOT NULL DEFAULT '';
CREATE INDEX idx_projects_status ON projects (status);
