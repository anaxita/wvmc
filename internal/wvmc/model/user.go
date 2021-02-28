package model

// User ...
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        int    `json:"role"`
	Password    string `json:"password,omitempty"`
	EncPassword string `json:"-"`
}
