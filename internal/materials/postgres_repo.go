package materials

import (
	"database/sql"
	"strconv"
	"strings"

	"vecai/internal/db"
)

// Postgres is the PostgreSQL-backed materials repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

const materialSelect = `SELECT m.*, s.name AS supplier_name, p.name AS project_name
	FROM materials m LEFT JOIN suppliers s ON s.id = m.supplier_id JOIN projects p ON p.id = m.project_id`

func (p *Postgres) ListMaterials(projectID, status, q any) ([]map[string]any, error) {
	sqlText := materialSelect
	args := []any{}
	conds := []string{}
	i := 1
	if projectID != nil {
		conds = append(conds, "m.project_id = $"+strconv.Itoa(i))
		args = append(args, projectID)
		i++
	}
	if status != nil {
		conds = append(conds, "m.status = $"+strconv.Itoa(i))
		args = append(args, status)
		i++
	}
	if q != nil {
		conds = append(conds, "m.name LIKE $"+strconv.Itoa(i))
		args = append(args, q)
		i++
	}
	if len(conds) > 0 {
		sqlText += " WHERE " + strings.Join(conds, " AND ")
	}
	sqlText += " ORDER BY m.name"
	return db.QueryMaps(p.database, sqlText, args...)
}

func (p *Postgres) GetMaterial(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, materialSelect+" WHERE m.id = $1", id)
}

func (p *Postgres) CreateMaterial(projectID int64, supplierID any, name, unit string, requiredQty, availableQty float64, status string, unitCost int64) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO materials (project_id, supplier_id, name, unit, required_qty, available_qty, used_qty, status, unit_cost) VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8) RETURNING id`,
		projectID, supplierID, name, unit, requiredQty, availableQty, status, unitCost).Scan(&id)
	return id, err
}

func (p *Postgres) InsertDelivery(materialID int64, quantity float64, note string, recordedBy int64) error {
	_, err := p.database.Exec(`INSERT INTO material_deliveries (material_id, quantity, note, recorded_by) VALUES ($1, $2, $3, $4)`,
		materialID, quantity, note, recordedBy)
	return err
}

func (p *Postgres) SetMaterialDelivery(id int64, available float64, status, lastDelivery string) error {
	_, err := p.database.Exec(`UPDATE materials SET available_qty = $1, status = $2, last_delivery = $3, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $4`,
		available, status, lastDelivery, id)
	return err
}

func (p *Postgres) ProjectExists(id int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = $1`, id).Scan(&c)
	return c > 0, err
}

func (p *Postgres) SupplierExists(id int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM suppliers WHERE id = $1`, id).Scan(&c)
	return c > 0, err
}

func (p *Postgres) TouchProject(projectID int64) error {
	_, err := p.database.Exec(`UPDATE projects SET updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $1`, projectID)
	return err
}

func (p *Postgres) LogActivity(projectID any, actorID any, name, typ, text string) {
	p.database.Exec(`INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES ($1, $2, $3, $4, $5)`,
		projectID, actorID, name, typ, text)
}

func (p *Postgres) ListSuppliers(category, q any) ([]map[string]any, error) {
	sqlText := "SELECT * FROM suppliers"
	args := []any{}
	conds := []string{}
	i := 1
	if category != nil {
		conds = append(conds, "category = $"+strconv.Itoa(i))
		args = append(args, category)
		i++
	}
	if q != nil {
		conds = append(conds, "(name LIKE $"+strconv.Itoa(i)+" OR category LIKE $"+strconv.Itoa(i+1)+")")
		args = append(args, q, q)
		i += 2
	}
	if len(conds) > 0 {
		sqlText += " WHERE " + strings.Join(conds, " AND ")
	}
	sqlText += " ORDER BY name"
	return db.QueryMaps(p.database, sqlText, args...)
}

func (p *Postgres) CreateSupplier(name, category, contact, phone, email, location string, rating float64) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO suppliers (name, category, contact, phone, email, location, rating) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		name, category, contact, phone, email, location, rating).Scan(&id)
	return id, err
}

func (p *Postgres) GetSupplier(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, "SELECT * FROM suppliers WHERE id = $1", id)
}
