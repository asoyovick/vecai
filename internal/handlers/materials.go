package handlers

import (
	"net/http"
	"strings"

	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

const materialSelect = `SELECT m.*, s.name AS supplier_name, p.name AS project_name
	FROM materials m LEFT JOIN suppliers s ON s.id = m.supplier_id JOIN projects p ON p.id = m.project_id`

func recomputeStatusAfterDelivery(available, required float64, current string) string {
	if required <= 0 {
		if current == "ordered" {
			return "ok"
		}
		return current
	}
	if current == "ordered" {
		return "ok"
	}
	if available/required >= 0.5 {
		return "ok"
	}
	return current
}

func (s *Server) ListMaterials(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	q := materialSelect
	var args []any
	var conds []string
	if v := r.URL.Query().Get("projectId"); v != "" {
		conds = append(conds, "m.project_id = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("status"); v != "" && v != "All" {
		conds = append(conds, "m.status = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("q"); v != "" {
		conds = append(conds, "m.name LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY m.name"
	rows, err := httpx.QueryMaps(s.DB, q, args...)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"materials": rows})
}

func (s *Server) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to add materials.") {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	projectID, err := httpx.AsNumber(body["projectId"], "Project", true, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Material name", true, 160)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	unit, err := httpx.AsString(body["unit"], "Unit", true, 24)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	requiredQty, err := httpx.AsNumber(body["requiredQty"], "Required quantity", false, 0, false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	availableQty, err := httpx.AsNumber(body["availableQty"], "Available quantity", false, 0, false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	var supplierID any
	if body["supplierId"] != nil && body["supplierId"] != "" {
		v, err := httpx.AsNumber(body["supplierId"], "Supplier", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		supplierID = int64(v)
	}
	unitCost, err := httpx.AsNumber(body["unitCost"], "Unit cost", false, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	status := "ok"
	if body["status"] != nil {
		status, err = httpx.AsOneOf(body["status"], "Status", []string{"ok", "low", "critical", "ordered"})
		if err != nil {
			httpx.Error(w, err)
			return
		}
	}
	var c int
	s.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", int64(projectID)).Scan(&c)
	if c == 0 {
		httpx.Error(w, httpx.Err(400, "Select a valid project."))
		return
	}
	if supplierID != nil {
		s.DB.QueryRow("SELECT COUNT(*) FROM suppliers WHERE id = ?", supplierID).Scan(&c)
		if c == 0 {
			httpx.Error(w, httpx.Err(400, "Select a valid supplier."))
			return
		}
	}
	res, err := s.DB.Exec(`INSERT INTO materials (project_id, supplier_id, name, unit, required_qty, available_qty, used_qty, status, unit_cost) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		int64(projectID), supplierID, name, unit, requiredQty, availableQty, status, int64(unitCost))
	if err != nil {
		httpx.Error(w, err)
		return
	}
	id, _ := res.LastInsertId()
	httpx.LogActivity(s.DB, int64(projectID), &httpx.Actor{ID: u.ID, Name: u.Name}, "material", "Added material “"+name+"” to tracking")
	httpx.TouchProject(s.DB, int64(projectID))
	row, err := httpx.QueryMap(s.DB, materialSelect+" WHERE m.id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"material": row})
}

func (s *Server) LogDelivery(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to log deliveries.") {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	quantity, err := httpx.AsNumber(body["quantity"], "Quantity", true, 0, false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if quantity <= 0 {
		httpx.Error(w, httpx.Err(400, "Quantity must be greater than zero."))
		return
	}
	note, _ := httpx.AsString(body["note"], "Note", false, 500)
	id := r.PathValue("id")
	material, err := httpx.QueryMap(s.DB, "SELECT * FROM materials WHERE id = ?", id)
	if err != nil {
		httpx.Error(w, httpx.Err(404, "Material not found."))
		return
	}
	if _, err := s.DB.Exec("INSERT INTO material_deliveries (material_id, quantity, note, recorded_by) VALUES (?, ?, ?, ?)", id, quantity, note, u.ID); err != nil {
		httpx.Error(w, err)
		return
	}
	newAvailable := material["available_qty"].(float64) + quantity
	newStatus := recomputeStatusAfterDelivery(newAvailable, material["required_qty"].(float64), material["status"].(string))
	today := timeNowDate()
	s.DB.Exec("UPDATE materials SET available_qty = ?, status = ?, last_delivery = ?, updated_at = datetime('now') WHERE id = ?", newAvailable, newStatus, today, id)
	httpx.LogActivity(s.DB, material["project_id"], &httpx.Actor{ID: u.ID, Name: u.Name}, "material", "Logged delivery: "+formatQty(quantity)+" "+material["unit"].(string)+" "+material["name"].(string))

	httpx.TouchProject(s.DB, material["project_id"].(int64))
	row, err := httpx.QueryMap(s.DB, materialSelect+" WHERE m.id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"material": row})
}
