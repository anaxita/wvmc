package entity

type ServerState string

const (
	ServerStateRunning   ServerState = "Running"
	ServerStateStopped   ServerState = "Off"
	ServerNetworkRunning ServerState = "LAN - Virtual Switch"
	ServerNetworkStopped ServerState = ""
)

// Server содержит модель сервера из БД
type Server struct {
	ID          int64   `json:"id" db:"id"`
	VMID        string  `json:"vmid" db:"vmid"`
	Name        string  `json:"name" db:"name"`
	HV          string  `json:"hv" db:"hv"`
	IP          string  `json:"ip" db:"ip"`
	OutAddr     string  `json:"out_addr" db:"out_addr"`
	Company     string  `json:"company" db:"company"`
	Description string  `json:"description" db:"description"`
	Memory      float64 `json:"memory" db:"memory"`
	Weight      int     `json:"weight" db:"weight"`
	State       string  `json:"state" db:"state"`
	Status      string  `json:"status" db:"status"`
	CpuLoad     int     `json:"cpu_load" db:"cpu_load"`
	CpuCores    int     `json:"cpu_cores" db:"cpu_cores"`
	Network     string  `json:"network" db:"network"`
	Backup      string  `json:"backup" db:"backup"`
	User        string  `json:"user" db:"user"`
	Password    string  `json:"password,omitempty" db:"password"`
}
