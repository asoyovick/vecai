package reports

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

// RegisterRoutes mounts daily report endpoints under /projects/:id/reports.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, access gin.HandlerFunc) {
	group := api.Group("/projects/:id/reports")
	group.Use(access)
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/summary", h.Summary)
		group.GET("/:reportId", h.Get)
		group.DELETE("/:reportId", h.Delete)
	}
}

type CreateRequest struct {
	ReportDate     string `json:"date" binding:"required"`
	CompletedWork  string `json:"completed_work" binding:"required"`
	WorkersPresent int    `json:"workers"`
	Remarks        string `json:"remarks"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateRequest
	if !api.BindJSON(c, &req) {
		return
	}

	report, err := h.service.Create(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Daily report created", report)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", items)
}

func (h *Handler) Summary(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", h.service.Summarize(items))
}

func (h *Handler) Get(c *gin.Context) {
	report, err := h.service.Get(c.Request.Context(), c.Param("id"), c.Param("reportId"))
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "", report)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"), c.Param("reportId"))
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "Daily report deleted", nil)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrReportNotFound):
		api.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		api.Fail(c, http.StatusBadRequest, err.Error())
	default:
		api.Fail(c, http.StatusInternalServerError, "internal server error")
	}
}
