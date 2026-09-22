package documents

import (
	"errors"
	"net/http"
	"path/filepath"

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

// RegisterRoutes mounts document endpoints under /projects/:id/documents.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	group := api.Group("/projects/:id/documents")
	{
		group.POST("", h.Upload)
		group.GET("", h.List)
		group.GET("/:documentId", h.Download)
		group.DELETE("/:documentId", h.Delete)
	}
}

func (h *Handler) Upload(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		api.Fail(c, http.StatusBadRequest, "multipart file field 'file' is required")
		return
	}

	doc, err := h.service.Upload(c.Request.Context(), c.Param("id"), userID, header, c.PostForm("category"))
	if err != nil {
		switch {
		case errors.Is(err, ErrFileTooLarge):
			api.Fail(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrInvalidInput):
			api.Fail(c, http.StatusBadRequest, err.Error())
		default:
			api.Fail(c, http.StatusInternalServerError, "failed to upload document")
		}
		return
	}

	api.OK(c, http.StatusCreated, "Document uploaded", doc)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", items)
}

// Download streams the stored file bytes to the client.
func (h *Handler) Download(c *gin.Context) {
	doc, err := h.service.Get(c.Request.Context(), c.Param("id"), c.Param("documentId"))
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			api.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	//
	file, err := h.service.OpenFile(doc)
	if err != nil {
		api.Fail(c, http.StatusNotFound, "stored file is missing")
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", `attachment; filename="`+filepath.Base(doc.Filename)+`"`)
	if doc.FileType != "" {
		c.Header("Content-Type", doc.FileType)
	}
	c.DataFromReader(http.StatusOK, doc.FileSize, doc.FileType, file, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"), c.Param("documentId"))
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			api.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "Document deleted", nil)
}
