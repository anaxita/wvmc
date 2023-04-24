package entity

const (
	UserRoleUser  = 0
	UserRoleAdmin = 1
)

// User ...
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	Role        int    `json:"role"`
	Password    string `json:"password,omitempty"`
	EncPassword string `json:"-"`
}
