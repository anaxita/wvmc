param (
    [string[]]$idList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$servers = 'DCSRVHV1','DCSRVHV2','DCSRVHV3','DCSRVHV4','DCSRVHV5','DCSRVHV6','DCSRVHV7','DCSRVHV8','DCSRVHV9','DCSRVHV10','DCSRVHV11','DCSRVHV12','DCSRVHV14','DCSRVHV15', 'DCSRVHVPITON', 'DCSRVHVTP', 'DCSRVHVTSG' | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_" Where-Object {$_.Id -in $Using:idList};
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
} -ThrottleLimit 5;

$servers