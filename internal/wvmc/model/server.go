package model

// Server содержит модель сервера из БД
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HV          string `json:"hv,omitempty"`
	IP          string `json:"ip,omitempty"`
	OutAddr     string `json:"out_addr,omitempty"`
	Company     string `json:"company,omitempty"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	Network     string `json:"network,omitempty"`
	User        string `json:"user,omitempty"`
	Password    string `json:"password,omitempty"`
}
