package project

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/api"
	"vecai/backend/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts project endpoints onto the versioned API group.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, access gin.HandlerFunc) {
	projects := api.Group("/projects")
	{
		projects.POST("", h.CreateProject)
		projects.GET("", h.ListProjects)

		detail := projects.Group("/:id")
		detail.Use(access)
		{
			detail.GET("", h.GetProjectByID)
			detail.PUT("", RequireProjectOwner(), h.UpdateProject)
			detail.DELETE("", RequireProjectOwner(), h.DeleteProject)

			detail.POST("/members", RequireProjectOwner(), h.InviteMember)
			detail.GET("/members", h.GetMembersByProject)
			detail.DELETE("/members/:memberId", RequireProjectOwner(), h.RemoveMember)

			detail.POST("/blueprints", h.AddBlueprint)
			detail.GET("/blueprints", h.GetBlueprintsByProject)

			detail.POST("/specs", h.AddSpec)
			detail.GET("/specs", h.GetSpecsByProject)
		}
	}
}

// ---- Request payloads ----

type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
	ClientName  string `json:"client_name"`
	Status      string `json:"status"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"expected_end_date"`
}

type UpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
	ClientName  *string `json:"client_name"`
	Status      *string `json:"status"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"expected_end_date"`
}

type AddBlueprintRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FilePath    string `json:"file_path" binding:"required"`
	Version     int    `json:"version"`
}

type AddSpecRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	SectionCode string `json:"section_code" binding:"required"`
}

type InviteMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// ---- Handlers ----

func (h *Handler) CreateProject(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateRequest
	if !api.BindJSON(c, &req) {
		return
	}

	created, err := h.service.CreateProject(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Project created successfully", created)
}

// ListProjects returns the authenticated user's projects.
func (h *Handler) ListProjects(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	projects, err := h.service.ListProjects(c.Request.Context(), userID)
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", projects)
}

func (h *Handler) GetProjectByID(c *gin.Context) {
	found, err := h.service.GetProjectByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "", found)
}

func (h *Handler) UpdateProject(c *gin.Context) {
	var req UpdateRequest
	if !api.BindJSON(c, &req) {
		return
	}

	updated, err := h.service.UpdateProject(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "Project updated successfully", updated)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	err := h.service.DeleteProject(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "Project deleted successfully", nil)
}

func (h *Handler) InviteMember(c *gin.Context) {
	var req InviteMemberRequest
	if !api.BindJSON(c, &req) {
		return
	}

	member, err := h.service.InviteMember(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Team member added", member)
}

func (h *Handler) GetMembersByProject(c *gin.Context) {
	members, err := h.service.GetMembersByProject(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", members)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	err := h.service.RemoveMember(c.Request.Context(), c.Param("id"), c.Param("memberId"))
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			api.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "Team member removed", nil)
}

func (h *Handler) AddBlueprint(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req AddBlueprintRequest
	if !api.BindJSON(c, &req) {
		return
	}

	blueprint, err := h.service.AddBlueprint(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Blueprint added", blueprint)
}

func (h *Handler) GetBlueprintsByProject(c *gin.Context) {
	blueprints, err := h.service.GetBlueprintsByProject(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", blueprints)
}

func (h *Handler) AddSpec(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req AddSpecRequest
	if !api.BindJSON(c, &req) {
		return
	}

	spec, err := h.service.AddSpec(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Spec added", spec)
}

func (h *Handler) GetSpecsByProject(c *gin.Context) {
	specs, err := h.service.GetSpecsByProject(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", specs)
}

// handleError maps domain errors onto the HTTP status codes in docs/api.md.
func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		api.Fail(c, http.StatusNotFound, "project not found")
	case errors.Is(err, ErrDuplicateProject):
		api.Fail(c, http.StatusConflict, "user is already a member of this project")
	case errors.Is(err, ErrInvalidInput):
		api.Fail(c, http.StatusBadRequest, err.Error())
	default:
		api.Fail(c, http.StatusInternalServerError, "internal server error")
	}
}
