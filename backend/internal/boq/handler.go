package boq

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/api"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts BOQ endpoints under /projects/:id/boq.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	group := api.Group("/projects/:id/boq")
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/compare", h.Compare)
		group.PUT("/:itemId", h.Update)
		group.DELETE("/:itemId", h.Delete)
	}
}

type ItemRequest struct {
	ItemName      string  `json:"item_name" binding:"required"`
	Quantity      float64 `json:"quantity" binding:"required,gt=0"`
	Unit          string  `json:"unit"`
	EstimatedCost float64 `json:"estimated_cost"`
	ActualCost    float64 `json:"actual_cost"`
}

func (h *Handler) Create(c *gin.Context) {
	var req ItemRequest
	if !api.BindJSON(c, &req) {
		return
	}

	item, err := h.service.Create(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusCreated, "BOQ item created", item)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", items)
}

// Compare returns estimated vs actual cost totals for the project.
func (h *Handler) Compare(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", h.service.Compare(items))
}

func (h *Handler) Update(c *gin.Context) {
	var req ItemRequest
	if !api.BindJSON(c, &req) {
		return
	}

	item, err := h.service.Update(c.Request.Context(), c.Param("id"), c.Param("itemId"), req)
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "BOQ item updated", item)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"), c.Param("itemId"))
	if err != nil {
		handleError(c, err)
		return
	}

	api.OK(c, http.StatusOK, "BOQ item deleted", nil)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrItemNotFound):
		api.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		api.Fail(c, http.StatusBadRequest, err.Error())
	default:
		api.Fail(c, http.StatusInternalServerError, "internal server error")
	}
}
