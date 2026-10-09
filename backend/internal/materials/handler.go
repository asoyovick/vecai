package materials

import (
	"errors"
	"net/http"
	"strconv"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Handler serves the materials and suppliers endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the materials endpoints on the mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/materials", h.ListMaterials)
	mux.HandleFunc("POST /api/materials", h.CreateMaterial)
	mux.HandleFunc("POST /api/materials/{id}/deliveries", h.LogDelivery)
	mux.HandleFunc("GET /api/suppliers", h.ListSuppliers)
	mux.HandleFunc("POST /api/suppliers", h.CreateSupplier)
}

func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID, status, q any
	if v := r.URL.Query().Get("projectId"); v != "" {
		projectID = v
	}
	if v := r.URL.Query().Get("status"); v != "" && v != "All" {
		status = v
	}
	if v := r.URL.Query().Get("q"); v != "" {
		q = "%" + v + "%"
	}
	materials, err := h.svc.ListMaterials(projectID, status, q)
	if err != nil {
		api.Error(w, 500, "Could not load materials.")
		return
	}
	api.JSON(w, 200, map[string]any{"materials": materials})
}

func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to add materials.") {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	material, err := h.svc.CreateMaterial(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"material": material})
}

func (h *Handler) LogDelivery(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to log deliveries.") {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid material id.")
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	material, err := h.svc.LogDelivery(u, id, body)
	if err != nil {
		if errors.Is(err, ErrMaterialNotFound) {
			api.Error(w, 404, err.Error())
			return
		}
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"material": material})
}

func (h *Handler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var category, q any
	if v := r.URL.Query().Get("category"); v != "" && v != "All" {
		category = v
	}
	if v := r.URL.Query().Get("q"); v != "" {
		q = "%" + v + "%"
	}
	suppliers, err := h.svc.ListSuppliers(category, q)
	if err != nil {
		api.Error(w, 500, "Could not load suppliers.")
		return
	}
	api.JSON(w, 200, map[string]any{"suppliers": suppliers})
}

func (h *Handler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to add suppliers.") {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	supplier, err := h.svc.CreateSupplier(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"supplier": supplier})
}
