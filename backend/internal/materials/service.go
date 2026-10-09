package materials

import (
	"errors"
	"fmt"
	"time"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Service contains the material and supplier business rules.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

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

func (s *Service) ListMaterials(projectID, status, q any) ([]map[string]any, error) {
	rows, err := s.repo.ListMaterials(projectID, status, q)
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

func (s *Service) CreateMaterial(u *auth.User, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to add materials.")
	}
	projectID, err := api.AsNumber(input["projectId"], "Project", true, 0, true)
	if err != nil {
		return nil, err
	}
	name, err := api.AsString(input["name"], "Material name", true, 160)
	if err != nil {
		return nil, err
	}
	unit, err := api.AsString(input["unit"], "Unit", true, 24)
	if err != nil {
		return nil, err
	}
	requiredQty, err := api.AsNumber(input["requiredQty"], "Required quantity", false, 0, false)
	if err != nil {
		return nil, err
	}
	availableQty, err := api.AsNumber(input["availableQty"], "Available quantity", false, 0, false)
	if err != nil {
		return nil, err
	}
	var supplierID any
	if input["supplierId"] != nil && input["supplierId"] != "" {
		v, err := api.AsNumber(input["supplierId"], "Supplier", false, 0, true)
		if err != nil {
			return nil, err
		}
		if ok, _ := s.repo.SupplierExists(int64(v)); !ok {
			return nil, errors.New("Select a valid supplier.")
		}
		supplierID = int64(v)
	}
	unitCost, err := api.AsNumber(input["unitCost"], "Unit cost", false, 0, true)
	if err != nil {
		return nil, err
	}
	status := "ok"
	if input["status"] != nil {
		status, err = api.AsOneOf(input["status"], "Status", Statuses)
		if err != nil {
			return nil, err
		}
	}
	if ok, _ := s.repo.ProjectExists(int64(projectID)); !ok {
		return nil, errors.New("Select a valid project.")
	}
	id, err := s.repo.CreateMaterial(int64(projectID), supplierID, name, unit, requiredQty, availableQty, status, int64(unitCost))
	if err != nil {
		return nil, err
	}
	s.repo.LogActivity(int64(projectID), u.ID, u.Name, "material", "Added material “"+name+"” to tracking")
	s.repo.TouchProject(int64(projectID))
	return s.repo.GetMaterial(id)
}

func (s *Service) LogDelivery(u *auth.User, id int64, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to log deliveries.")
	}
	quantity, err := api.AsNumber(input["quantity"], "Quantity", true, 0, false)
	if err != nil {
		return nil, err
	}
	if quantity <= 0 {
		return nil, errors.New("Quantity must be greater than zero.")
	}
	note, _ := api.AsString(input["note"], "Note", false, 500)
	material, err := s.repo.GetMaterial(id)
	if err != nil {
		return nil, ErrMaterialNotFound
	}
	if err := s.repo.InsertDelivery(id, quantity, note, u.ID); err != nil {
		return nil, err
	}
	newAvailable := toF(material["available_qty"]) + quantity
	newStatus := recomputeStatusAfterDelivery(newAvailable, toF(material["required_qty"]), material["status"].(string))
	today := time.Now().UTC().Format("2006-01-02")
	if err := s.repo.SetMaterialDelivery(id, newAvailable, newStatus, today); err != nil {
		return nil, err
	}
	s.repo.LogActivity(material["project_id"], u.ID, u.Name, "material",
		fmt.Sprintf("Logged delivery: %v %s %s", quantity, material["unit"], material["name"]))
	s.repo.TouchProject(material["project_id"].(int64))
	return s.repo.GetMaterial(id)
}

func (s *Service) ListSuppliers(category, q any) ([]map[string]any, error) {
	rows, err := s.repo.ListSuppliers(category, q)
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

func (s *Service) CreateSupplier(u *auth.User, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to add suppliers.")
	}
	name, err := api.AsString(input["name"], "Supplier name", true, 160)
	if err != nil {
		return nil, err
	}
	category, err := api.AsString(input["category"], "Category", true, 80)
	if err != nil {
		return nil, err
	}
	contact, _ := api.AsString(input["contact"], "Contact person", false, 120)
	phone, _ := api.AsString(input["phone"], "Phone", false, 40)
	email, _ := api.AsString(input["email"], "Email", false, 200)
	location, _ := api.AsString(input["location"], "Location", false, 160)
	rating, err := api.AsNumber(input["rating"], "Rating", false, 0, false)
	if err != nil {
		return nil, err
	}
	if rating > 5 {
		return nil, errors.New("Rating must be between 0 and 5.")
	}
	id, err := s.repo.CreateSupplier(name, category, contact, phone, email, location, rating)
	if err != nil {
		return nil, err
	}
	return s.repo.GetSupplier(id)
}

func toF(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case string:
		var f float64
		fmt.Sscanf(x, "%f", &f)
		return f
	}
	return 0
}
