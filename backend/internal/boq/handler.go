package boq

import (
	"errors"
	"net/http"
	"strconv"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Handler serves the BOQ endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the BOQ endpoints on the mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/boq", h.List)
	mux.HandleFunc("POST /api/boq", h.Create)
	mux.HandleFunc("PATCH /api/boq/{id}", h.Patch)
	mux.HandleFunc("DELETE /api/boq/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID any
	if v := r.URL.Query().Get("projectId"); v != "" {
		projectID = v
	}
	items, err := h.svc.List(projectID)
	if err != nil {
		api.Error(w, 500, "Could not load the BOQ.")
		return
	}
	api.JSON(w, 200, map[string]any{"items": items})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to edit the BOQ.") {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	item, err := h.svc.Create(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"item": item})
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to edit the BOQ.") {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid BOQ item id.")
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	item, err := h.svc.Patch(u, id, body)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			api.Error(w, 404, err.Error())
			return
		}
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 200, map[string]any{"item": item})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to edit the BOQ.") {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid BOQ item id.")
		return
	}
	if err := h.svc.Delete(u, id); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			api.Error(w, 404, err.Error())
			return
		}
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 200, map[string]any{"ok": true})
}
