package materials

// Repository is the persistence contract for materials and suppliers.
type Repository interface {
	ListMaterials(projectID, status, q any) ([]map[string]any, error)
	GetMaterial(id int64) (map[string]any, error)
	CreateMaterial(projectID int64, supplierID any, name, unit string, requiredQty, availableQty float64, status string, unitCost int64) (int64, error)
	InsertDelivery(materialID int64, quantity float64, note string, recordedBy int64) error
	SetMaterialDelivery(id int64, available float64, status, lastDelivery string) error
	ProjectExists(id int64) (bool, error)
	SupplierExists(id int64) (bool, error)
	TouchProject(projectID int64) error
	LogActivity(projectID any, actorID any, name, typ, text string)
	ListSuppliers(category, q any) ([]map[string]any, error)
	CreateSupplier(name, category, contact, phone, email, location string, rating float64) (int64, error)
	GetSupplier(id int64) (map[string]any, error)
}
