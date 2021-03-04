package model

// Server содержит модель сервера из БД
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IP4      string `json:"ip4"`
	HV       string `json:"hv"`
	User     string `json:"user"`
	Password string `json:"password"`
}
