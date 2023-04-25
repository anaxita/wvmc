package entity

const (
	UserRoleUser  = 0
	UserRoleAdmin = 1
)

// User ...
type User struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Email       string `json:"email" db:"email"`
	Company     string `json:"company" db:"company"`
	Role        int    `json:"role" db:"role"`
	Password    string `json:"password,omitempty" db:"password"`
	EncPassword string `json:"-" db:"-"`
}
