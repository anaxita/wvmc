package entity

type Command = string

const (
	CommandStartPower     Command = "start_power"
	CommandStopPower      Command = "stop_power"
	CommandStopPowerForce Command = "stop_power_force"

	CommandStartNetwork Command = "start_network"
	CommandStopNetwork  Command = "stop_network"
)

type ServerState string

const (
	ServerStateRunning   ServerState = "Running"
	ServerStateStopped   ServerState = "Off"
	ServerNetworkRunning ServerState = "LAN - Virtual Switch"
	ServerNetworkStopped ServerState = ""
)

// Server содержит модель сервера из БД
type Server struct {
	ID     string `json:"id" db:"id"`
	Title  string `json:"name" db:"name"`
	HV     string `json:"hv" db:"hv"`
	State  string `json:"state" db:"state"`
	Status string `json:"status" db:"status"`
}
