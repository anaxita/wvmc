package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anaxita/wvmc/internal/entity"
)

type VM struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Network string `json:"network,omitempty"`
	HV      string `json:"HV,omitempty"`
}

type WinServices struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"status"`
	User        string `json:"user"`
}

type WinVolume struct {
	Letter     string  `json:"disk_letter"`
	SpaceTotal float32 `json:"space_total"`
	SpaceFree  float32 `json:"space_free"`
}

type WinProcess struct {
	SessionID int    `json:"session_id"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CPULoad   int    `json:"cpu_load"`
	Memory    int    `json:"memory"`
}

type WinRDPSesion struct {
	SessionID int          `json:"session_id"`
	UserName  string       `json:"user_name"`
	State     string       `json:"state"`
	Processes []WinProcess `json:"processes"`
}

// Control содержит структуру, которая реализует интерфейс Commander
type Control struct {
}

func NewControlService() *Control {
	return &Control{}
}

func (s *Control) ControlServer(ctx context.Context, server entity.Server, command entity.Command) (err error) {
	// TODO add ct to all functions

	switch command {
	case entity.CommandStartPower:
		_, err = s.startServer(server)
	case entity.CommandStopPower:
		_, err = s.stopServer(server)
	case entity.CommandStopPowerForce:
		_, err = s.stopServerForce(server)
	case entity.CommandStartNetwork:
		_, err = s.startServerNetwork(server)
	case entity.CommandStopNetwork:
		_, err = s.stopServerNetwork(server)
	default:
		return fmt.Errorf("%w: unknown command %s", entity.ErrValidate, command)
	}

	return err
}

// Servers получает статус работы всех ВМ servers
func (s *Control) Servers(ctx context.Context) ([]entity.Server, error) {
	hvs := os.Getenv("HV_LIST")

	scriptPath := "./powershell/GetVmForAdmins.ps1"

	out, err := s.run(scriptPath, "-hvList", hvs)
	if err != nil {
		return nil, err
	}

	var servers []entity.Server

	if err = json.Unmarshal(out, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}

// stopServer выключает сервер
func (s *Control) stopServer(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Title %s -ComputerName '%s'", server.Title, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// stopServerForce принудительно выключает сервер
func (s *Control) stopServerForce(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Title %s -Force -ComputerName '%s'", server.Title, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// startServer включает сервер
func (s *Control) startServer(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -Title %s -ComputerName '%s'", server.Title, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// startServerNetwork включает сеть на сервере
func (s *Control) startServerNetwork(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Connect-VMNetworkAdapter -VMName %s -SwitchName \"DMZ - Virtual Switch\" -ComputerName '%s'",
		server.Title, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// stopServerNetwork выключает сеть на сервере
func (s *Control) stopServerNetwork(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Disconnect-VMNetworkAdapter -VMName %s -ComputerName '%s'", server.Title,
		server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Control) getServerData(server entity.Server, hv string, name string) (entity.Server,
	error) {
	scriptPath := "./powershell/GetVmByHvAndName.ps1"

	out, err := s.run(scriptPath, "-hv", hv, "-name", name)
	if err != nil {
		return server, err
	}

	if err = json.Unmarshal(out, &server); err != nil {
		return server, err
	}

	return server, nil
}

// getServerServices получает список служб сервера
func (s *Control) getServerServices(ip, user, password string) ([]WinServices, error) {
	var services []WinServices
	scriptPath := "./powershell/getServerServices.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s'", scriptPath, ip, user, password)

	out, err := s.run(args)
	if err != nil {
		return services, err
	}

	if err = json.Unmarshal(out, &services); err != nil {
		return services, err
	}

	return services, nil
}

// startWinService включает службу сервера
func (s *Control) startWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/StartService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// stopWinService выключает службу сервера
func (s *Control) stopWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/StopService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// restartWinService переззагружает службу сервера
func (s *Control) restartWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/RestartService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// getDiskFreeSpace получает информацию о свободном мсесте на дисках
func (s *Control) getDiskFreeSpace(ip, user, password string) ([]WinVolume, error) {
	var disks []WinVolume
	scriptPath := "./powershell/getDiskFreeSpace.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s'", scriptPath, ip, user, password)
	out, err := s.run(args)
	if err != nil {
		return disks, err
	}

	if err = json.Unmarshal(out, &disks); err != nil {
		return disks, err
	}

	return disks, nil
}

// getProcesses получает информацию о процессах (диспетчер задач)
func (s *Control) getProcesses(ip, user, password string) ([]WinRDPSesion, error) {
	processes := []WinRDPSesion{}
	scriptPath := "./powershell/getProcesses.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s'", scriptPath, ip, user, password)

	out, err := s.run(args)
	if err != nil {
		return processes, err
	}

	// check if not sessions
	if string(out) == "" {
		return processes, nil
	}

	if err = json.Unmarshal(out, &processes); err != nil {
		return processes, err
	}

	return processes, nil
}

// stoptWinProcess force stop process by id
func (s *Control) stoptWinProcess(ip, user, password string, id int) ([]byte, error) {
	scriptPath := "./powershell/StopProcess.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -id '%d'", scriptPath, ip, user, password, id)
	return s.run(args)
}

// disconnectRDPUser close RDP user session
func (s *Control) disconnectRDPUser(ip, user, password string, sessionID int) ([]byte,
	error) {
	scriptPath := "./powershell/disconnectRDPUser.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -id %d", scriptPath, ip, user, password,
		sessionID)
	return s.run(args)
}

func (s *Control) run(args ...string) ([]byte, error) {
	command := strings.Join(args, " ")

	e := exec.Command("pwsh", "-NoLogo", "-Mta", "-NoProfile", "-NonInteractive", "-Command",
		command)

	out, err := e.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
