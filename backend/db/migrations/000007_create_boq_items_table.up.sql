-- Bills of Quantities: estimated vs actual cost per line item.
CREATE TABLE IF NOT EXISTS boq_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    item_name VARCHAR(255) NOT NULL,
    quantity DECIMAL(14,3) NOT NULL CHECK (quantity > 0),
    unit VARCHAR(50),
    estimated_cost DECIMAL(14,2) NOT NULL DEFAULT 0 CHECK (estimated_cost >= 0),
    actual_cost DECIMAL(14,2) NOT NULL DEFAULT 0 CHECK (actual_cost >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_boq_items_project_id ON boq_items(project_id);
