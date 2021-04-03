package model

// Server содержит модель сервера из БД
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HV          string `json:"hv"`
	IP          string `json:"ip"`
	OutAddr     string `json:"out_addr"`
	Company     string `json:"company"`
	Description string `json:"description"`
	State       string `json:"state"`
	Network     string `json:"network"`
	User        string `json:"user,omitempty"`
	Password    string `json:"password,omitempty"`
}
