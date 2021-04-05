package control

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// VM описывает свойства виртуальной машины, который можно получить с помощью комманд данного пакета
type VM struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Network string `json:"network,omitempty"`
	HV      string `json:"HV,omitempty"`
}

// Commander описывает метод который запускает команду powershell,возвращает вывод и ошибку
type Commander interface {
	run(args ...string) ([]byte, error)
}

// Command содержит методы Run для запуска powershell команд
type Command struct{}

// run запускает команду powershell,возвращает вывод и ошибку
func (c *Command) run(args ...string) ([]byte, error) {
	command := strings.Join(args, " ")

	e := exec.Command("pwsh", "-NoLogo", "-Mta", "-NoProfile", "-NonInteractive", "-Command", command)
	logit.Log("COMMAND", e.Args)

	out, err := e.Output()
	logit.Log("out: ", string(out))
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

// GetServersDataForUsers получает статус работы и сети ВМ servers по их Name
func (s *ServerService) GetServersDataForUsers(servers []model.Server) ([]model.Server, error) {
	var uniqHVs = make(map[string]bool)
	var ids string
	var hvs string
	var vms = make([]model.Server, 0)
	var scriptPath = "./powershell/GetVmForUsers.ps1"

	for _, v := range servers {
		uniqHVs[v.HV] = false

		if ids == "" {
			ids = fmt.Sprintf("'%s'", v.ID)
		} else {
			ids = fmt.Sprintf("%s, '%s'", ids, v.ID)
		}
	}

	for k := range uniqHVs {
		if hvs == "" {
			hvs = fmt.Sprintf("'%v'", k)
		} else {
			hvs = fmt.Sprintf("%s, '%v'", hvs, k)
		}
	}

	out, err := s.commander.run(scriptPath, "-hvList", hvs, "-idList", ids)
	if err != nil {
		logit.Log("Ошибка powershell ", err)
		return vms, err
	}

	if err = json.Unmarshal(out, &vms); err != nil {
		logit.Log("Ошибка json unmarshal ", err)
		return vms, err
	}
	return vms, nil
}

// GetServersDataForAdmins получает статус работы всех ВМ servers
func (s *ServerService) GetServersDataForAdmins() ([]model.Server, error) {
	hvs := os.Getenv("HV_LIST")

	scriptPath := "./powershell/GetVMForAdmins.ps1"

	out, err := s.commander.run(scriptPath, "-hvList", hvs)
	if err != nil {
		return nil, err
	}

	var servers []model.Server

	if err = json.Unmarshal(out, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}

// GetServerDataForAdmins получает статус работы всех ВМ servers
func (s *ServerService) GetServerDataForAdmins(hv string) ([]model.Server, error) {
	script := `$hvList = 'DCSRVHV1','DCSRVHV2';
	$idList = '15332a09-a1fa-42e2-97e3-35f19e0f3a86', '5f08e450-4342-452f-af0a-4e5594ac9dbe', '2f89e03b-e72d-4867-9ffb-44dd06cc6163', 'bbe86300-1329-4526-b108-7b780c9c3f57';
	$allServers = $hvList | ForEach-Object -Parallel {
		$servers = Get-VM -ComputerName "$_" | Where-Object {$_.Id -in $Using:idList};
		if ($null -ne $servers) {
	
			foreach ($s in $servers)
			{
				$state = $s.State;
				if ($state -eq 2) {
					$state = "Running";
				} else {
					$state = "Off";
				}
				
				[pscustomobject]@{
					"id" = $s.Id;
					"name" = $s.Name;
					"state" = $state;
					"hv" = $s.ComputerName;
				};
				
			}
		}
	}
	$allServers | ConvertTo-Json -AsArray -Compress;`

	// TODO для получения ВСЕХ серверов юзать этот скрипт
	_ = `$result = New-Object System.Collections.Arraylist;
    $servers = Get-VM -ComputerName $hvList;
foreach ($s in $servers)
{
        $state = $s.State;
    
        if ($state -eq 2) {
            $state = "Running";
        } else {
            $state = "Off";
        }
    
        $vm = @{
            "id" = $s.Id;
			"name" = $s.Name;
            "state" = $state;
			"hv" = $s.ComputerName;
        };

        $result.Add($vm) | Out-Null
    }
    
    $result | ConvertTo-Json -AsArray -Compress;`

	// command := fmt.Sprintf("$hvList = %s;  %s", hv, script)

	out, err := s.commander.run(script)
	if err != nil {
		return nil, err
	}

	var vms []model.Server

	if err = json.Unmarshal(out, &vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// StopServer выключает сервер
func (s *ServerService) StopServer(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Name '%s' -ComputerName '%s'", server.Name, server.HV)
	return s.commander.run(command)
}

// StopServerForce принудительно выключает сервер
func (s *ServerService) StopServerForce(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -Name '%s' -Force -ComputerName '%s'", server.Name, server.HV)
	return s.commander.run(command)
}

// StartServer включает сервер
func (s *ServerService) StartServer(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -Name '%s' -ComputerName '%s'", server.Name, server.HV)
	return s.commander.run(command)
}

// StartServerNetwork включает сеть на сервере
func (s *ServerService) StartServerNetwork(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Connect-VMNetworkAdapter -VMName %s -SwitchName \"DMZ - Virtual Switch\" -ComputerName '%s'", server.Name, server.HV)
	return s.commander.run(command)
}

// StopServerNetwork выключает сеть на сервере
func (s *ServerService) StopServerNetwork(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Disconnect-VMNetworkAdapter -VMName %s -ComputerName '%s'", server.Name, server.HV)
	return s.commander.run(command)
}

// UpdateAllServersInfo обновляет информацию по всем серверам в БД
func (s *ServerService) UpdateAllServersInfo() ([]model.Server, error) {
	hvs := os.Getenv("HV_LIST")

	scriptPath := "./powershell/GetVmToDB.ps1"

	out, err := s.commander.run(scriptPath, "-hvList", hvs)
	if err != nil {
		return nil, err
	}

	var servers []model.Server

	if err = json.Unmarshal(out, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
