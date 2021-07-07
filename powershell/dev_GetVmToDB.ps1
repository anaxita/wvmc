param (
    [string[]]$hvList 
)

$nameList = 'ServerTwo', 'VMBitrix_dev.kmsys.ru_off', 'VM_TestCentOS_off', 'SRV_IPZinchenko_off';

$result = $hvList | ForEach-Object -Parallel {
    Get-VM  -ComputerName "$_" | Where-Object {$_.Name -in $Using:nameList} | ForEach-Object -Parallel {

        $ip = "No data";
        
        if ($_.State -eq 2) {
            $networkAdapter = $_ | Get-VMNetworkAdapter;
            $ip4 = $networkAdapter.IPAddresses;
            if ($null -ne $ip4[0]) {
                $ip = $ip4 -join ', ';
            };
        }

        [pscustomobject]@{
            "id" = $_.Id;
            "name" = $_.Name;
            "ip" = $ip;
            "hv" = $_.ComputerName;
        }

    } -ThrottleLimit 5;
} -ThrottleLimit 5;
$result | ConvertTo-Json -AsArray -Compress;