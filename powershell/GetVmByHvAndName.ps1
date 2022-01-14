param (
    [string]$hv,
    [string]$name
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
    $vm = Get-VM -ComputerName $hv -Name $name
            $state = $vm.State;
            if ($state -eq 2) {
                $state = "Running"
            } else {
                $state = "Off"
            }
            
            $data = @{
                "vmid" = $vm.Id
                "name" = $vm.Name
                "state" = $state
                "cpu_load" = $vm.CPUUsage
                "cpu_cores" = $vm.ProcessorCount
                "weight" = ($vm | Get-VMProcessor).RelativeWeight
                "description" = $vm.Description
                "memory" = [math]::Round(($vm.MemoryStartup / 1GB), 0)
                "network" = ($vm | Get-VMNetworkAdapter).SwitchName
                "hv" = $vm.ComputerName
            }

$data | ConvertTo-Json -Compress