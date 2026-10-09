package auth

import "strings"

// initialsFor derives two-letter initials from a full name.
func initialsFor(name string) string {
	parts := strings.Fields(name)
	out := ""
	for i, p := range parts {
		if i >= 2 {
			break
		}
		if len(p) > 0 {
			out += strings.ToUpper(p[:1])
		}
	}
	if out == "" {
		out = "U"
	}
	return out
}
