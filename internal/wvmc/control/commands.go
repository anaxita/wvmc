package control

import (
	"fmt"
	"os/exec"
)

// VM описывает свойства виртуальной машины, который можно получить с помощью комманд данного пакета
type VM struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"power_status"`
	Network string `json:"network_status"`
	HV      string `json:"HV"`
}

// Commander описывает метод который запускает команду powershell,возвращает вывод и ошибку
type Commander interface {
	Run(string) ([]byte, error)
}

// Command содержит методы Run для запуска powershell команд
type Command struct{}

// ServerService содержит структуру, которая реализует интерфейс Commander
type ServerService struct {
	Commander Commander
}

// Run запускает команду powershell,возвращает вывод и ошибку
func (c *Command) Run(command string) ([]byte, error) {
	e := exec.Command("powershell", command)
	out, _ := e.Output()
	return out, nil
}

// NewServerService ...
func NewServerService(c Commander) *ServerService {
	return &ServerService{
		Commander: c,
	}
}

// GetServerStatus получает статус работы и сети ВМ
func (s *ServerService) GetServerStatus(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Get-Server-Status -Name %s", serverID)
	return s.Commander.Run(command)
}

// StopServer выключает сервер
func (s *ServerService) StopServer(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Get-Server-Status -Name %s", serverID)
	return s.Commander.Run(command)
}

// StopServerForce принудительно выключает сервер
func (s *ServerService) StopServerForce(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Get-Server-Status -Name %s", serverID)
	return s.Commander.Run(command)
}

// StartServer включает сервер
func (s *ServerService) StartServer(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Get-Server-Status -Name %s", serverID)
	return s.Commander.Run(command)
}
