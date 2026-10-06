// Package project owns projects, their budget and their tasks.
package project

import "errors"

// Statuses are the allowed project lifecycle stages.
var Statuses = []string{"Tendering", "Design", "Foundation", "Construction", "Roofing", "Finishing", "Handover", "Completed"}

// TaskStatuses are the allowed task statuses.
var TaskStatuses = []string{"not-started", "in-progress", "blocked", "completed"}

// TaskPriorities are the allowed task priorities.
var TaskPriorities = []string{"high", "medium", "low"}

var (
	ErrProjectNotFound = errors.New("Project not found.")
	ErrTaskNotFound    = errors.New("Task not found.")
	ErrNoFields        = errors.New("No fields to update.")
)

// FieldUpdate is one column assignment for a partial update.
type FieldUpdate struct {
	Column string
	Value  any
}
