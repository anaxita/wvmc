param (
    [string[]]$hvList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$servers = $hvList | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_";
    if ($null -ne $vms) {
        foreach ($vm in $vms)
        {
            $state = $vm.State;
            if ($state -eq 2) {
                $state = "Running";
            } else {
                $state = "Off";
            }
            
            [pscustomobject]@{
                "id" = $vm.Id;
                "name" = $vm.Name;
                "state" = $state;
                "hv" = $vm.ComputerName;
            } ;
            
        }
    }
} -ThrottleLimit 3;

$servers |ConvertTo-Json -AsArray -Compress