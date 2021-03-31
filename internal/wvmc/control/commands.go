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
	run(string) ([]byte, error)
}

// Command содержит методы Run для запуска powershell команд
type Command struct{}

// run запускает команду powershell,возвращает вывод и ошибку
func (c *Command) run(command string) ([]byte, error) {
	e := exec.Command("pwsh", "-Command", command)
	// e := exec.Command("pwsh", "./powershell/test.ps1")
	logit.Info("Выполняем команду", command)
	out, err := e.Output()
	logit.Info(string(out))
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
	var hvs []string
	var names []string
	var allNames string
	var allHV string

	for _, v := range servers {
		hvs = append(hvs, v.HV)
		names = append(names, v.Name)
	}

	allHV = strings.Join(hvs, ",")

	nameList, err := json.Marshal(names)
	if err != nil {
		logit.Log("Ошибка маршала", err)
	}

	allNames = strings.Replace(string(nameList), "[", "", 1)
	allNames = strings.Replace(allNames, "]", "", 1)
	logit.Log(allNames)

	script := ` | ForEach-Object -Parallel {
    $state = $_.State;
    
    if ($state -eq 2) {
        $state = "Running";
    } else {
        $state = "Off";
    }

    $network = ($_ | Get-VMNetworkAdapter).SwitchName;
    if ($network -eq "DMZ - Virtual Switch") {
        $network = "Running";
    } else {
        $network = "Off";
    }

    [pscustomobject]@{
        "id" = $_.Id;
		"name" = $_.Name;
        "state" = $state;
        "network" = $network;
    };
};

$result | ConvertTo-Json -AsArray -Compress;`
	vms := make([]model.Server, 0)

	command := fmt.Sprintf("$nameList = %s; $result = Get-VM -Name $nameList -ComputerName %s %s", allNames, allHV, script)

	out, err := s.commander.run(command)
	logit.Log("out:", string(out))
	if err != nil {
		return vms, err
	}

	if err = json.Unmarshal(out, &vms); err != nil {
		return vms, err
	}
	return vms, nil
}

// GetServersDataForAdmins получает статус работы всех ВМ servers
func (s *ServerService) GetServersDataForAdmins() ([]VM, error) {
	script := `$result = Get-VM -ComputerName $hvList | ForEach-Object -Parallel {
    $state = $_.State;

    if ($state -eq 2) {
        $state = "Running";
    } else {
        $state = "Off";
    }

    [pscustomobject]@{
        "id" = $_.Id;
		"name" = $_.Name;
        "state" = $state;
		"hv" = $_.ComputerName;
    };
} -ThrottleLimit 5;

$result | ConvertTo-Json -AsArray -Compress;`

	scriptsCimSessinons := `Get-CimSession`

	command := fmt.Sprintf("$hvList = %s; %s", os.Getenv("HV_LIST"), script)
	logit.Log(command)
	out, err := s.commander.run(scriptsCimSessinons)
	if err != nil {
		return nil, err
	}

	var vms []VM

	if err = json.Unmarshal(out, &vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// GetServerDataForAdmins получает статус работы всех ВМ servers
func (s *ServerService) GetServerDataForAdmins(hv string) ([]model.Server, error) {
	script := `$result = New-Object System.Collections.Arraylist;
    $servers = Get-VM -ComputerName $hvList | Where-Object {$_.Id -in '15332a09-a1fa-42e2-97e3-35f19e0f3a86', '5f08e450-4342-452f-af0a-4e5594ac9dbe', '2f89e03b-e72d-4867-9ffb-44dd06cc6163', 'bbe86300-1329-4526-b108-7b780c9c3f57'};
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

	command := fmt.Sprintf("$hvList = %s;  %s", hv, script)
	// logit.Log(command)
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

// StopServer выключает сервер
func (s *ServerService) StopServer(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -ID '%s' -ComputerName '%s'", server.ID, server.HV)
	return s.commander.run(command)
}

// StopServerForce принудительно выключает сервер
func (s *ServerService) StopServerForce(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Stop-VM -ID '%s' -Force -ComputerName '%s'", server.ID, server.HV)
	return s.commander.run(command)
}

// StartServer включает сервер
func (s *ServerService) StartServer(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID '%s' -ComputerName '%s'", server.ID, server.HV)
	return s.commander.run(command)
}

// StartServerNetwork включает сеть на сервере
func (s *ServerService) StartServerNetwork(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID '%s' -ComputerName '%s'", server.ID, server.HV)
	return s.commander.run(command)
}

// StopServerNetwork выключает сеть на сервере
func (s *ServerService) StopServerNetwork(server model.Server) ([]byte, error) {
	command := fmt.Sprintf("Start-VM -ID '%s' -ComputerName '%s'", server.ID, server.HV)
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
	$result | ConvertTo-Json -AsArray -Compress;`
	return s.commander.run(command)
}
