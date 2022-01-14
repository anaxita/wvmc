package model

type ServerState string

const (
	ServerStateRunning   ServerState = "Running"
	ServerStateStopped   ServerState = "Off"
	ServerNetworkRunning ServerState = "LAN - Virtual Switch"
	ServerNetworkStopped ServerState = ""
)

// Server содержит модель сервера из БД
type Server struct {
	ID          int64   `json:"id"`
	VMID        string  `json:"vmid"`
	Name        string  `json:"name"`
	HV          string  `json:"hv"`
	IP          string  `json:"ip"`
	OutAddr     string  `json:"out_addr"`
	Company     string  `json:"company"`
	Description string  `json:"description"`
	Memory      float64 `json:"memory"`
	Weight      int     `json:"weight"`
	State       string  `json:"state"`
	Status      string  `json:"status"`
	CpuLoad     int     `json:"cpu_load"`
	CpuCores    int     `json:"cpu_cores"`
	Network     string  `json:"network"`
	Backup      string  `json:"backup"`
	User        string  `json:"user"`
	Password    string  `json:"password,omitempty"`
}
