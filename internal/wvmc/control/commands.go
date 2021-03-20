package control

import (
	"fmt"
	"os/exec"
	"strings"
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
	run(string) ([]byte, error)
}

// Command содержит методы Run для запуска powershell команд
type Command struct{}

// run запускает команду powershell,возвращает вывод и ошибку
func (c *Command) run(command string) ([]byte, error) {
	e := exec.Command("pwsh", "-Command", command)
	out, err := e.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerService содержит структуру, которая реализует интерфейс Commander
type ServerService struct {
	commander Commander
}

// NewServerService ...
func NewServerService(c Commander) *ServerService {
	return &ServerService{
		commander: c,
	}
}

// GetServerStatus получает статус работы и сети ВМ servers по их ID
func (s *ServerService) GetServerStatus(servers []string) ([]byte, error) {
	script := `
$result = New-Object System.Collections.Arraylist;
foreach ($s in $servers) {
	$power_status = $s.State
	if ($power_status -eq 3) {
		$network_status = 3
	} else {
		$network_status = ($s | Get-VMNetworkAdapter).Status
	}

	$vm = @{
		"id" = $s.Id
		"name" = $s.Name
		"power" = $power_status
		"network" = $network_status[0]
	}

	$result.Add($vm) | Out-Null;
}
$result | ConvertTo-Json;
	`
	command := fmt.Sprintf("$servers = Get-VM -ID %s ; %s", strings.Join(servers, ","), script)

	return s.commander.run(command)
}

// StopServer выключает сервер
func (s *ServerService) StopServer(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -ID %s", serverID)
	return s.commander.run(command)
}

// StopServerForce принудительно выключает сервер
func (s *ServerService) StopServerForce(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -ID %s -Force", serverID)
	return s.commander.run(command)
}

// StartServer включает сервер
func (s *ServerService) StartServer(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID %s", serverID)
	return s.commander.run(command)
}

// StartServerNetwork включает сеть на сервере
func (s *ServerService) StartServerNetwork(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID %s", serverID)
	return s.commander.run(command)
}

// StopServerNetwork выключает сеть на сервере
func (s *ServerService) StopServerNetwork(serverID string) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID %s", serverID)
	return s.commander.run(command)
}

// UpdateAllServersInfo обновляет информацию по всем серверам в БД
func (s *ServerService) UpdateAllServersInfo() ([]byte, error) {
	command := `$hvList = 'DCSRVHV1','DCSRVHV2','DCSRVHV3','DCSRVHV4','DCSRVHV5','DCSRVHV6','DCSRVHV7','DCSRVHV8','DCSRVHV9','DCSRVHV10','DCSRVHV11','DCSRVHV12','DCSRVHV14','DCSRVHV15', 'DCSRVHVPITON', 'DCSRVHVTP' , 'DCSRVHVTSG';
	$servers = Get-VM -ComputerName $hvList;
	$result = New-Object System.Collections.Arraylist;
	foreach ($s in $servers)
	{
		$networkAdapter = $s | Get-VMNetworkAdapter;
		# $network = $networkAdapter.SwitchName;
		$ip = "no data";
		$ip4 = $networkAdapter.IPAddresses
		if ($null -ne $ip4[0]) {
			$ip = $ip4 -join ', ';
		};
		$vm = @{
			"id" = $s.VMId;
			"name" = $s.VMName;
			"ip" = $ip;
			# "notes" = $s.NOTES;
			# "state" = $s.State;
			# "network" = $network;
			# "ram" = $s.MemoryStartup/1MB;
			"hv" = $s.ComputerName;
		};
		$result.add($vm) | Out-Null;
	};
	$result | ConvertTo-Json -AsArray;`
	return s.commander.run(command)
}
