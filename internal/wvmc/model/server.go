package model

// Server содержит модель сервера из БД
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"`
	OutAddr     string `json:"out_addr"`
	HV          string `json:"hv"`
	Company     string `json:"company"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	Network     string `json:"network,omitempty"`
	User        string `json:"user,omitempty"`
	Password    string `json:"password,omitempty"`
}
