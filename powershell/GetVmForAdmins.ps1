param (
    [string[]]$hvList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
# $servers = $hvList | ForEach-Object -Parallel {
#     $vms = Get-VM -ComputerName "$_";
#     if ($Null -ne $vms) {
#         foreach ($vm in $vms) {
#             $state = $vm.State;
#             $networkAdapter = $vm | Get-VMNetworkAdapter;
#             if ($state -eq 2) {
#                 $ip4 = $networkAdapter.IPAddresses
#                 $state = "Running";
#             }
#             else {
#                 $state = "Off";
#             }
            
#             [pscustomobject]@{
#                 "id"      = $vm.Id;
#                 "name"    = $vm.Name;
#                 "state"   = $state;
#                 "network" = [string]$networkAdapter.SwitchName;
#                 "status"  = $vm.Status;
#                 "cpu"     = $vm.CPUUsage;
#                 "hv"      = $vm.ComputerName;
#                 "ip"      = $ip4[0];
#             } ;
            
#         }
#     }
# } -ThrottleLimit 30;

$servers =  $hvList | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_";

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