package boq

import (
	"errors"

	"vecai/internal/api"
	"vecai/internal/auth"
)

var ErrItemNotFound = errors.New("BOQ item not found.")

// Service holds the BOQ business rules.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(projectID any) ([]map[string]any, error) {
	rows, err := s.repo.ListItems(projectID)
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

func (s *Service) Create(u *auth.User, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to edit the BOQ.")
	}
	projectID, err := api.AsNumber(input["projectId"], "Project", true, 0, true)
	if err != nil {
		return nil, err
	}
	if ok, _ := s.repo.ProjectExists(int64(projectID)); !ok {
		return nil, errors.New("Select a valid project.")
	}
	itemNo, _ := api.AsString(input["itemNo"], "Item number", false, 60)
	description, err := api.AsString(input["description"], "Description", true, 400)
	if err != nil {
		return nil, err
	}
	unit, _ := api.AsString(input["unit"], "Unit", false, 24)
	if unit == "" {
		unit = "pcs"
	}
	quantity, err := api.AsNumber(input["quantity"], "Quantity", true, 0, false)
	if err != nil {
		return nil, err
	}
	unitRate, err := api.AsNumber(input["unitRate"], "Unit rate", true, 0, true)
	if err != nil {
		return nil, err
	}
	id, err := s.repo.CreateItem(int64(projectID), itemNo, description, unit, quantity, int64(unitRate))
	if err != nil {
		return nil, err
	}
	s.repo.LogActivity(int64(projectID), u.ID, u.Name, "project", "Added BOQ item “"+description+"”")
	s.repo.TouchProject(int64(projectID))
	return s.repo.GetItem(id)
}

func (s *Service) Patch(u *auth.User, id int64, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to edit the BOQ.")
	}
	existing, err := s.repo.GetItem(id)
	if err != nil {
		return nil, ErrItemNotFound
	}
	updates := []FieldUpdate{}
	if v, ok := input["itemNo"]; ok {
		s2, err := api.AsString(v, "Item number", false, 60)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"item_no", s2})
	}
	if v, ok := input["description"]; ok {
		s2, err := api.AsString(v, "Description", true, 400)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"description", s2})
	}
	if v, ok := input["unit"]; ok {
		s2, err := api.AsString(v, "Unit", true, 24)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"unit", s2})
	}
	if v, ok := input["quantity"]; ok {
		n, err := api.AsNumber(v, "Quantity", true, 0, false)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"quantity", n})
	}
	if v, ok := input["unitRate"]; ok {
		n, err := api.AsNumber(v, "Unit rate", true, 0, true)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"unit_rate", int64(n)})
	}
	if len(updates) == 0 {
		return nil, errors.New("No fields to update.")
	}
	if err := s.repo.UpdateItem(id, updates); err != nil {
		return nil, err
	}
	s.repo.TouchProject(existing["project_id"].(int64))
	return s.repo.GetItem(id)
}

func (s *Service) Delete(u *auth.User, id int64) error {
	if u.Role == "viewer" {
		return errors.New("You do not have permission to edit the BOQ.")
	}
	existing, err := s.repo.GetItem(id)
	if err != nil {
		return ErrItemNotFound
	}
	if err := s.repo.DeleteItem(id); err != nil {
		return err
	}
	s.repo.TouchProject(existing["project_id"].(int64))
	return nil
}
