package model

// User ...
type User struct {
	ID          string `json:"id,string"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	Role        int    `json:"role"`
	Password    string `json:"password,omitempty"`
	EncPassword string `json:"-"`
}
