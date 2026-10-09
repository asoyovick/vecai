// Package reports provides the workspace dashboard, the activity feed,
// daily site reports and the project-intelligence assistant.
package reports

import (
	"errors"
	"strconv"
	"strings"
)

// parseMaterials validates the materialsUsed array of a daily site report.
func parseMaterials(raw any) ([]map[string]any, error) {
	arr, ok := raw.([]any)
	if !ok {
		return nil, nil
	}
	var out []map[string]any
	for _, item := range arr {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		quantity, err := asAmount(entry["quantity"])
		if err != nil || quantity <= 0 {
			return nil, errors.New("Enter a quantity greater than zero for “" + name + "”.")
		}
		unit, _ := entry["unit"].(string)
		if len(unit) > 24 {
			unit = unit[:24]
		}
		if len(name) > 120 {
			name = name[:120]
		}
		out = append(out, map[string]any{"name": name, "quantity": quantity, "unit": strings.TrimSpace(unit)})
	}
	return out, nil
}

func asAmount(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case string:
		return strconv.ParseFloat(x, 64)
	}
	return 0, errors.New("not a number")
}
