param (
    [string[]]$hvList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$servers =  $hvList | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_" | Where-Object {$_.ReplicationState -lt 2};

    if ($null -eq $vms) {
        continue
    }

        $vms | ForEach-Object -Parallel {
            $state = $_.State;
            $networkAdapter = $_ | Get-VMNetworkAdapter;
            $ip = ''

            $ip4 = $networkAdapter.IPAddresses
            if ($ip4 -gt 0) {
                $ip = $ip4[0]
            }
            if ($state -eq 2) {
                $state = "Running";
            }
            else {
                $state = "Off";
            }

            [PSCustomObject]@{
                "vmid"      = $_.Id;
                "name"    = $_.Name;
                "state"   = $state;
                "network" = [string]$networkAdapter.SwitchName;
                "status"  = $_.Status;
                "cpu"     = $_.CPUUsage;
                "hv"      = $_.ComputerName;
                "ip"      = $ip;
            }
        
    }
} -ThrottleLimit 30;


$servers | ConvertTo-Json -AsArray -Compress