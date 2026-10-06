package auth

import (
	"net/http"
	"strings"
	"time"

	"vecai/internal/api"
)

// Handler serves the auth and user-management endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	email, err := api.AsString(body["email"], "Email", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	password, err := api.AsString(body["password"], "Password", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	u, token, err := h.svc.Login(email, password)
	if err != nil {
		api.Error(w, 401, "Incorrect email or password.")
		return
	}
	http.SetCookie(w, SessionCookieHeader(token, time.Now().Add(sessionTTL)))
	api.JSON(w, 200, map[string]any{"user": u})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	name, err := api.AsString(body["name"], "Full name", true, 120)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	email, err := api.AsString(body["email"], "Email", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	password, err := api.AsString(body["password"], "Password", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	jobTitle, _ := api.AsString(body["jobTitle"], "Job title", false, 120)
	u, token, err := h.svc.Register(name, strings.ToLower(email), password, jobTitle)
	if err != nil {
		switch err {
		case ErrEmailTaken:
			api.Error(w, 409, err.Error())
		default:
			api.Error(w, 400, err.Error())
		}
		return
	}
	http.SetCookie(w, SessionCookieHeader(token, time.Now().Add(sessionTTL)))
	api.JSON(w, 201, map[string]any{"user": u})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(SessionCookie); err == nil && c.Value != "" {
		h.svc.Logout(c.Value)
	}
	http.SetCookie(w, ClearSessionCookieHeader())
	api.JSON(w, 200, map[string]any{"ok": true})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u := RequireUser(w, r)
	if u == nil {
		return
	}
	api.JSON(w, 200, map[string]any{"user": u})
}

func (h *Handler) ListTeam(w http.ResponseWriter, r *http.Request) {
	if RequireUser(w, r) == nil {
		return
	}
	members, assignments, err := h.svc.Team()
	if err != nil {
		api.Error(w, 500, "Could not load the team.")
		return
	}
	api.JSON(w, 200, map[string]any{"members": members, "assignments": assignments})
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	u := RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	name, err := api.AsString(body["name"], "Full name", true, 120)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	email, err := api.AsString(body["email"], "Email", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	email = strings.ToLower(email)
	jobTitle, err := api.AsString(body["jobTitle"], "Job title", true, 120)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	role := "member"
	if body["role"] != nil {
		role, err = api.AsOneOf(body["role"], "Role", []string{"member", "viewer"})
		if err != nil {
			api.Error(w, 400, err.Error())
			return
		}
	}
	var projectID *int64
	if body["projectId"] != nil && body["projectId"] != "" {
		v, err := api.AsNumber(body["projectId"], "Project", false, 0, true)
		if err != nil {
			api.Error(w, 400, err.Error())
			return
		}
		pid := int64(v)
		projectID = &pid
	}
	phone, _ := api.AsString(body["phone"], "Phone", false, 40)
	if !api.EmailRe.MatchString(email) {
		api.Error(w, 400, "Enter a valid email address.")
		return
	}
	member, tempPassword, err := h.svc.InviteMember(u, name, email, jobTitle, role, phone, projectID)
	if err != nil {
		if err == ErrEmailTaken {
			api.Error(w, 409, err.Error())
			return
		}
		api.Error(w, 500, "Could not create the account.")
		return
	}
	api.JSON(w, 201, map[string]any{"member": member, "temporaryPassword": tempPassword})
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	u := RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	name, err := api.AsString(body["name"], "Name", true, 120)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	email, err := api.AsString(body["email"], "Email", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	phone, _ := api.AsString(body["phone"], "Phone", false, 40)
	jobTitle, _ := api.AsString(body["jobTitle"], "Job title", false, 120)
	row, err := h.svc.UpdateProfile(u, name, strings.ToLower(email), phone, jobTitle)
	if err != nil {
		api.Error(w, statusFor(err), err.Error())
		return
	}
	api.JSON(w, 200, map[string]any{"user": row.publicMap()})
}

func statusFor(err error) int {
	if strings.Contains(err.Error(), "already in use") {
		return 409
	}
	if strings.Contains(err.Error(), "valid email") || strings.Contains(err.Error(), "required") {
		return 400
	}
	return 400
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u := RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	current, err := api.AsString(body["currentPassword"], "Current password", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	newPassword, err := api.AsString(body["newPassword"], "New password", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	if err := h.svc.ChangePassword(u, current, newPassword); err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	http.SetCookie(w, ClearSessionCookieHeader())
	api.JSON(w, 200, map[string]any{"ok": true})
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	u := RequireUser(w, r)
	if u == nil {
		return
	}
	prefs, _ := h.svc.GetPreferences(u.ID)
	api.JSON(w, 200, map[string]any{"preferences": prefs})
}

func (h *Handler) PatchPreferences(w http.ResponseWriter, r *http.Request) {
	u := RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	prefs, err := h.svc.PatchPreferences(u.ID, body)
	if err != nil {
		api.Error(w, 500, "Could not save preferences.")
		return
	}
	api.JSON(w, 200, map[string]any{"preferences": prefs})
}

func (h *Handler) GetOrganisation(w http.ResponseWriter, r *http.Request) {
	if RequireUser(w, r) == nil {
		return
	}
	row, err := h.svc.GetOrganisation()
	if err != nil {
		api.Error(w, 500, "Could not load organisation.")
		return
	}
	api.JSON(w, 200, map[string]any{"organisation": map[string]any{
		"name": row["name"], "businessType": row["business_type"], "country": row["country"], "currency": row["currency"],
	}})
}

func (h *Handler) UpdateOrganisation(w http.ResponseWriter, r *http.Request) {
	if RequireRole(w, r, "admin") == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	name, err := api.AsString(body["name"], "Organisation name", true, 160)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	businessType, err := api.AsString(body["businessType"], "Business type", true, 80)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	country, err := api.AsString(body["country"], "Country", true, 80)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	currency, err := api.AsString(body["currency"], "Currency", true, 80)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	if err := h.svc.UpdateOrganisation(name, businessType, country, currency); err != nil {
		api.Error(w, 500, "Could not save organisation.")
		return
	}
	api.JSON(w, 200, map[string]any{"organisation": map[string]string{"name": name, "businessType": businessType, "country": country, "currency": currency}})
}
