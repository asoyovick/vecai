package materials

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

// RegisterRoutes mounts material tracking endpoints under /projects/:id/materials.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, access gin.HandlerFunc) {
	group := api.Group("/projects/:id/materials")
	group.Use(access)
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/summary", h.Summary)
		group.PUT("/:materialId", h.Update)
		group.DELETE("/:materialId", h.Delete)
	}
}

type CreateRequest struct {
	MaterialName string  `json:"material_name" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
	Unit         string  `json:"unit" binding:"required"`
	UnitCost     float64 `json:"unit_cost"`
	Supplier     string  `json:"supplier"`
	DeliveryDate string  `json:"delivery_date"`
}

type UpdateRequest struct {
	MaterialName *string  `json:"material_name"`
	Quantity     *float64 `json:"quantity"`
	Unit         *string  `json:"unit"`
	UnitCost     *float64 `json:"unit_cost"`
	Supplier     *string  `json:"supplier"`
	DeliveryDate *string  `json:"delivery_date"`
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

	material, err := h.service.Create(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "Material delivery recorded", material)
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

func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if !api.BindJSON(c, &req) {
		return
	}

	material, err := h.service.Update(c.Request.Context(), c.Param("id"), c.Param("materialId"), req)
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "Material record updated", material)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"), c.Param("materialId"))
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "Material record deleted", nil)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrMaterialNotFound):
		api.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		api.Fail(c, http.StatusBadRequest, err.Error())
	default:
		api.Fail(c, http.StatusInternalServerError, "internal server error")
	}
}
