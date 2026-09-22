package auth

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

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if !api.BindJSON(c, &req) {
		return
	}

	if err := ValidateRegister(req); err != nil {
		api.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			api.Fail(c, http.StatusConflict, "an account with this email already exists")
			return
		}
		api.Fail(c, http.StatusInternalServerError, "registration failed")
		return
	}

	api.OK(c, http.StatusCreated, "Account created successfully", resp)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			api.Fail(c, http.StatusUnauthorized, err.Error())
			return
		}
		if errors.Is(err, ErrAccountNotVerified) {
			api.Fail(c, http.StatusForbidden, err.Error())
			return
		}
		if isValidationError(err) {
			api.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		api.Fail(c, http.StatusInternalServerError, "authentication failed")
		return
	}

	api.OK(c, http.StatusOK, "Login successful", resp)
}

// Verify handles POST /api/v1/auth/verify
func (h *Handler) Verify(c *gin.Context) {
	var req VerifyRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.service.Verify(c.Request.Context(), req)
	if err != nil {
		api.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	api.OK(c, http.StatusOK, "Account verified", resp)
}

// Me handles GET /api/v1/users/me
func (h *Handler) Me(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.service.Profile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			api.Fail(c, http.StatusNotFound, "user not found")
			return
		}
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "", user)
}

// UpdateMe handles PUT /api/v1/users/me
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		api.Fail(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req UpdateProfileRequest
	if !api.BindJSON(c, &req) {
		return
	}

	user, err := h.service.UpdateProfile(c.Request.Context(), userID, req.Fullname, req.Phone)
	if err != nil {
		if errors.Is(err, ErrFullNameEmpty) {
			api.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			api.Fail(c, http.StatusNotFound, "user not found")
			return
		}
		api.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}

	api.OK(c, http.StatusOK, "Profile updated", user)
}

// Logout handles POST /api/v1/auth/logout. JWTs are stateless, so the client
// discards the token; the endpoint exists to keep the API contract symmetric.
func (h *Handler) Logout(c *gin.Context) {
	api.OK(c, http.StatusOK, "Logged out", nil)
}

func isValidationError(err error) bool {
	return errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrPasswordTooWeak) ||
		errors.Is(err, ErrFullNameEmpty)
}
