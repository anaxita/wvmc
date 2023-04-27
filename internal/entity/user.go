package entity

type CtxUserKey struct{}

type UserRole = int

const (
	UserRoleUser UserRole = iota
	UserRoleAdmin
)

type User struct {
	ID       int64    `json:"id" db:"id"`
	Name     string   `json:"name" db:"name"`
	Email    string   `json:"email" db:"email"`
	Company  string   `json:"company" db:"company"`
	Role     UserRole `json:"role" db:"role"`
	Password string   `json:"-" db:"password"`
}
