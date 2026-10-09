// Package materials tracks project materials, deliveries and suppliers.
package materials

import "errors"

var (
	ErrMaterialNotFound = errors.New("Material not found.")
)

var Statuses = []string{"ok", "low", "critical", "ordered"}
