package control

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	// e := exec.Command("pwsh", "-NoLogo", "-Mta", "-NoProfile", "-NonInteractive", "-File", "./powershell/update_servers.ps1")
	// e := exec.Command("pwsh", "./powershell/test.ps1")
	logit.Log("COMMAND", e.Args)

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

// GetServersDataForUsers получает статус работы и сети ВМ servers по их Name
func (s *ServerService) GetServersDataForUsers(servers []model.Server) ([]model.Server, error) {
	start := time.Now()
	logit.Info("Начало", start)
	var hvs = make(map[string]bool)
	var names []string
	var allNames string
	var allHV string

	for _, v := range servers {
		hvs[v.HV] = false
		names = append(names, v.Name)
	}

	logit.Log("карта ", hvs)

	for k := range hvs {
		allHV = fmt.Sprintf("'%s',", k)
	}

	logit.Log("после цикла ", allHV)
	allHV = strings.TrimRight(allHV, ",")
	logit.Log("после трима ", allHV)

	nameList, err := json.Marshal(names)
	if err != nil {
		logit.Log("Ошибка маршала", err)
	}

	allNames = strings.Replace(string(nameList), "[", "", 1)
	allNames = strings.Replace(allNames, "]", "", 1)
	allNames = strings.Replace(allNames, ",", ", ", -1)

	script := `$allServers = $hvList | ForEach-Object -Parallel {
		$servers = Get-VM -ComputerName "$_" | Where-Object {$_.Name -in $Using:nameList};
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
	} -ThrottleLimit 3;
	$allServers | ConvertTo-Json -AsArray -Compress;`
	vms := make([]model.Server, 0)

	command := fmt.Sprintf("$nameList = %s ; $hvList = %s ; %s", allNames, allHV, script)
	end := time.Since(start)
	logit.Info("Конец", end)

	logit.Log(command)
	out, err := s.commander.run(command)

	if err != nil {
		return vms, err
	}

	if err = json.Unmarshal(out, &vms); err != nil {
		return vms, err
	}
	return vms, nil
}

// GetServersDataForAdmins получает статус работы всех ВМ servers
func (s *ServerService) GetServersDataForAdmins() ([]model.Server, error) {
	command := `$hvList = 'DCSRVHV1','DCSRVHV2';
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

	out, err := s.commander.run(command)
	if err != nil {
		return nil, err
	}

	var vms []model.Server

	if err = json.Unmarshal(out, &vms); err != nil {
		return nil, err
	}

	return vms, nil
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
	hvs := fmt.Sprintf("$hvList = %s;", os.Getenv("HV_LIST"))

	script := `$servers = Get-VM -ComputerName $hvList;
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
	$result | ConvertTo-Json -AsArray -Compress;`

	command := fmt.Sprintf("%s %s", hvs, script)
	out, err := s.commander.run(command)
	if err != nil {
		return nil, err
	}

	var servers []model.Server

	if err = json.Unmarshal(out, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
