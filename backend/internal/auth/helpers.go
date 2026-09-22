package auth

import "strings"

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
