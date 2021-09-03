param (
    [string[]]$hvList 
)
[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$result = $hvList | ForEach-Object -Parallel {
    Get-VM  -ComputerName "$_" | ForEach-Object -Parallel {

        $ip = "";
        
        if ($_.State -eq 2) {
            $networkAdapter = $_ | Get-VMNetworkAdapter;
            $ip4 = $networkAdapter.IPAddresses
            if ($null -ne $ip4[0]) {
                $ip = $ip4[0]
            };
        }

        [pscustomobject]@{
            "id" = $_.Id;
            "name" = $_.Name;
            "ip" = $ip;
            "hv" = $_.ComputerName;
        }

    } -ThrottleLimit 10;
} -ThrottleLimit 10;
$result | ConvertTo-Json -AsArray -Compress;