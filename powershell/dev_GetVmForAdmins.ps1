param (
    [string[]]$hvList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$nameList = 'ServerTwo' , 'VMBitrix_dev.kmsys.ru_off', 'VM_TestCentOS_off';
$hvList =  'DCSRVHV2';
$servers = $hvList | ForEach-Object -Parallel {
    $vms = Get-VM -ComputerName "$_" | Where-Object {$_.Name -in $Using:nameList};
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
                "status" = $vm.Status;
                "cpu" = $vm.CPUUsage;
                "hv" = $vm.ComputerName;
            } ;
            
        }
    }
} -ThrottleLimit 3;

$servers |ConvertTo-Json -AsArray -Compress