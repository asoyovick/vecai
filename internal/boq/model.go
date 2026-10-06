package boq

// FieldUpdate is one column assignment for a partial update.
type FieldUpdate struct {
	Column string
	Value  any
}
