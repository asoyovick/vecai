// Package api holds the shared HTTP concerns for every domain:
// consistent JSON responses, the JSON content-type guard, CORS and panic
// recovery.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// JSON writes a successful JSON response.
func JSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		JSON(w, 500, map[string]string{"error": "failed to encode response"})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	w.Write(body)
}

// Error writes a consistent {"error": msg} response.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Decode reads a JSON request body with a 1 MB cap.
func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			return ErrTooLarge
		}
		return ErrBadJSON
	}
	return nil
}

var (
	ErrTooLarge = errors.New("Request body too large.")
	ErrBadJSON  = errors.New("Invalid JSON body.")
)

// HandleDecodeError maps api.Decode errors to responses.
func HandleDecodeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrTooLarge) {
		Error(w, 413, err.Error())
	} else {
		Error(w, 400, err.Error())
	}
	return true
}

// Mutations must arrive as JSON or multipart, a lightweight CSRF guard.
func JSONGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			ct := r.Header.Get("Content-Type")
			isJSON := strings.Contains(ct, "application/json")
			isMultipart := strings.Contains(ct, "multipart/form-data")
			bodyless := ct == "" && r.ContentLength <= 0
			if !isJSON && !isMultipart && !bodyless {
				Error(w, 415, "Unsupported content type.")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// IsSecureRequest reports whether the request arrived over TLS, directly or
// through a proxy that set X-Forwarded-Proto.
func IsSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// Recover converts panics into 500s instead of dropped connections.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				Error(w, 500, "Something went wrong on the server. Please try again.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ── Shared field validation ──────────────────────────────────────────────────

func AsString(value any, field string, required bool, max int) (string, error) {
	if value == nil {
		if required {
			return "", fmt.Errorf("%s is required.", field)
		}
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string.", field)
	}
	s = strings.TrimSpace(s)
	if required && s == "" {
		return "", fmt.Errorf("%s is required.", field)
	}
	if max > 0 && len(s) > max {
		return "", fmt.Errorf("%s must be at most %d characters.", field, max)
	}
	return s, nil
}

func AsNumber(value any, field string, required bool, min float64, integer bool) (float64, error) {
	if value == nil || value == "" {
		if required {
			return 0, fmt.Errorf("%s is required.", field)
		}
		return 0, nil
	}
	var n float64
	switch v := value.(type) {
	case float64:
		n = v
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be a number.", field)
		}
		n = parsed
	default:
		return 0, fmt.Errorf("%s must be a number.", field)
	}
	if integer && n != float64(int64(n)) {
		return 0, fmt.Errorf("%s must be a whole number.", field)
	}
	if n < min {
		return 0, fmt.Errorf("%s must be at least %v.", field, min)
	}
	return n, nil
}

func AsOneOf(value any, field string, allowed []string) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be one of: %s.", field, strings.Join(allowed, ", "))
	}
	for _, a := range allowed {
		if s == a {
			return s, nil
		}
	}
	return "", fmt.Errorf("%s must be one of: %s.", field, strings.Join(allowed, ", "))
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func AsDate(value any, field string, required bool) (*string, error) {
	if value == nil || value == "" {
		if required {
			return nil, fmt.Errorf("%s is required.", field)
		}
		return nil, nil
	}
	s, ok := value.(string)
	if !ok || !dateRe.MatchString(s) {
		return nil, fmt.Errorf("%s must be a date (YYYY-MM-DD).", field)
	}
	return &s, nil
}

var EmailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
