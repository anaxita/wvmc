param (
    [string[]]$hvList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$nameList = 'ServerTwo' , 'VMBitrix_dev.kmsys.ru_off', 'VM_TestCentOS_off';
$hvList =  'DCSRVHV2', 'DCSRVHV7' ;

$servers =  $hvList | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_" | Where-Object {$_.Name -in $Using:nameList};

    if ($null -eq $vms) {
        return $false
    }

        $vms | ForEach-Object -Parallel {
            $state = $_.State;
            $networkAdapter = $_ | Get-VMNetworkAdapter;
            $ip = ''

            $ip4 = $networkAdapter.IPAddresses
            if ($null -ne $ip4[0]) {
                $ip = $ip4[0]
            }
            if ($state -eq 2) {
                $state = "Running";
            }
            else {
                $state = "Off";
            }

            [PSCustomObject]@{
                "id"      = $_.Id;
                "name"    = $_.Name;
                "state"   = $state;
                "network" = [string]$networkAdapter.SwitchName;
                "status"  = $vm.Status;
                "cpu"     = $_.CPUUsage;
                "hv"      = $_.ComputerName;
                "ip"      = $ip;
            }
        
    }
} -ThrottleLimit 30;


$servers | ConvertTo-Json -AsArray -Compress