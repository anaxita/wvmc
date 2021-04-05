param (
    [string[]]$hvList,
    [string[]]$idList
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$result = $hvList | ForEach-Object -Parallel {
    Get-VM  -ComputerName "$_" | Where-Object {$_.Id -in $Using:idList} | ForEach-Object -Parallel {
        $network = ($_ | Get-VMNetworkAdapter).SwitchName;
        if ($network -eq "DMZ - Virtual Switch") {
            $network = "Running";
        } else {
            $network = "Off";
        }

        $state = $_.State
        if ($state -eq 2) {
            $state = "Running";
        } else {
            $state = "Off";
        }

        [pscustomobject]@{
            "id" = $_.Id;
            "name" = $_.Name;
            "network" = $network;
            "hv" = $_.ComputerName;
        }

    } -ThrottleLimit 5;
} -ThrottleLimit 5;
$result | ConvertTo-Json -AsArray -Compress;