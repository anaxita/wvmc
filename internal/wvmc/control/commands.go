package control

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anaxita/wvmc/internal/wvmc/cache"

	"github.com/anaxita/wvmc/internal/wvmc/entity"
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

// Commander описывает метод который запускает команду powershell,возвращает вывод и ошибку
type Commander interface {
	run(args ...string) ([]byte, error)
}

// Command содержит методы Run для запуска powershell команд
type Command struct{}

// run запускает команду powershell,возвращает вывод и ошибку // TODO DEPRECATED
func (c *Command) run(args ...string) ([]byte, error) {
	command := strings.Join(args, " ")

	e := exec.Command("pwsh", "-NoLogo", "-Mta", "-NoProfile", "-NonInteractive", "-Command",
		command)

	out, err := e.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Service содержит структуру, которая реализует интерфейс Commander
type Service struct {
	cache *cache.CacheService
}

func NewServerService(cache *cache.CacheService) *Service {
	return &Service{cache: cache}
}

func (s *Service) run(args ...string) ([]byte, error) {
	command := strings.Join(args, " ")

	e := exec.Command("pwsh", "-NoLogo", "-Mta", "-NoProfile", "-NonInteractive", "-Command",
		command)

	out, err := e.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// GetServersDataForUsers получает статус работы и сети ВМ servers по их Name
func (s *Service) GetServersDataForUsers(servers []entity.Server) ([]entity.Server, error) {
	var uniqHVs = make(map[string]bool)
	var ids string
	var hvs string
	var vms = make([]entity.Server, 0)
	var scriptPath = "./powershell/GetVmForUsers.ps1"

	for _, v := range servers {
		uniqHVs[v.HV] = false

		if ids == "" {
			ids = fmt.Sprintf("'%s'", v.VMID)
		} else {
			ids = fmt.Sprintf("%s, '%s'", ids, v.VMID)
		}
	}

	for k := range uniqHVs {
		if hvs == "" {
			hvs = fmt.Sprintf("'%v'", k)
		} else {
			hvs = fmt.Sprintf("%s, '%v'", hvs, k)
		}
	}

	out, err := s.run(scriptPath, "-hvList", hvs, "-idList", ids)
	if err != nil {
		return vms, err
	}

	if err = json.Unmarshal(out, &vms); err != nil {
		return vms, err
	}
	return vms, nil
}

// GetServersDataForAdmins получает статус работы всех ВМ servers
func (s *Service) GetServersDataForAdmins() ([]entity.Server, error) {
	if s.cache.Servers() != nil {
		return s.cache.Servers(), nil
	}

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

	s.cache.SetServers(servers)

	return servers, nil
}

// StopServer выключает сервер
func (s *Service) StopServer(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Name %s -ComputerName '%s'", server.Name, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	s.cache.SetServerState(server, entity.ServerStateStopped)
	return out, nil
}

// StopServerForce принудительно выключает сервер
func (s *Service) StopServerForce(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Name %s -Force -ComputerName '%s'", server.Name, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	s.cache.SetServerState(server, entity.ServerStateStopped)
	return out, nil
}

// StartServer включает сервер
func (s *Service) StartServer(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -Name %s -ComputerName '%s'", server.Name, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	s.cache.SetServerState(server, entity.ServerStateRunning)
	return out, nil
}

// StartServerNetwork включает сеть на сервере
func (s *Service) StartServerNetwork(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Connect-VMNetworkAdapter -VMName %s -SwitchName \"DMZ - Virtual Switch\" -ComputerName '%s'",
		server.Name, server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	s.cache.SetServerNetwork(server, entity.ServerNetworkRunning)
	return out, nil
}

// StopServerNetwork выключает сеть на сервере
func (s *Service) StopServerNetwork(server entity.Server) ([]byte, error) {
	command := fmt.Sprintf("Disconnect-VMNetworkAdapter -VMName %s -ComputerName '%s'", server.Name,
		server.HV)
	out, err := s.run(command)
	if err != nil {
		return nil, err
	}

	s.cache.SetServerNetwork(server, entity.ServerNetworkStopped)
	return out, nil
}

func (s *Service) GetServerData(server entity.Server, hv string, name string) (entity.Server,
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

// GetServerServices получает список служб сервера
func (s *Service) GetServerServices(ip, user, password string) ([]WinServices, error) {
	var services []WinServices
	scriptPath := "./powershell/GetServerServices.ps1"
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

// StartWinService включает службу сервера
func (s *Service) StartWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/StartService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// StopWinService выключает службу сервера
func (s *Service) StopWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/StopService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// RestartWinService переззагружает службу сервера
func (s *Service) RestartWinService(ip, user, password, serviceName string) ([]byte, error) {
	scriptPath := "./powershell/RestartService.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -name '%s'", scriptPath, ip, user, password,
		serviceName)
	return s.run(args)
}

// GetServerServices получает информацию о свободном мсесте на дисках
func (s *Service) GetDiskFreeSpace(ip, user, password string) ([]WinVolume, error) {
	var disks []WinVolume
	scriptPath := "./powershell/GetDiskFreeSpace.ps1"
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

// GetProcesses получает информацию о процессах (диспетчер задач)
func (s *Service) GetProcesses(ip, user, password string) ([]WinRDPSesion, error) {
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

// StoptWinProcess force stop process by id
func (s *Service) StoptWinProcess(ip, user, password string, id int) ([]byte, error) {
	scriptPath := "./powershell/StopProcess.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -id '%d'", scriptPath, ip, user, password, id)
	return s.run(args)
}

// DisconnectRDPUser close RDP user session
func (s *Service) DisconnectRDPUser(ip, user, password string, sessionID int) ([]byte,
	error) {
	scriptPath := "./powershell/DisconnectRDPUser.ps1"
	args := fmt.Sprintf("%s -ip %s -u '%s' -p '%s' -id %d", scriptPath, ip, user, password,
		sessionID)
	return s.run(args)
}
