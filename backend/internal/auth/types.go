package auth

// User is the authenticated identity shared with every domain.
type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JobTitle string `json:"jobTitle"`
	Initials string `json:"initials"`
	Color    string `json:"color"`
	Phone    string `json:"phone"`
}

// UserRow is the stored user record including the password hash.
type UserRow struct {
	User
	PasswordHash string
}

// publicMap renders the non-sensitive fields of a stored user as an API map.
func (r *UserRow) publicMap() map[string]any {
	return map[string]any{
		"id": r.ID, "email": r.Email, "name": r.Name, "role": r.Role,
		"jobTitle": r.JobTitle, "initials": r.Initials, "color": r.Color, "phone": r.Phone,
	}
}
