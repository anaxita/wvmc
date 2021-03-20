package model

// Server содержит модель сервера из БД
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	HV       string `json:"hv"`
	Company  string `json:"company"`
	User     string `json:"user"`
	Password string `json:"password"`
}
