[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")

$hvList = 'DCSRVHV10', 'DCSRVHV12', 'DCSRVHVPITON', 'DCSRVHV8', 'DCSRVHV3', 'DCSRVHV1', 'DCSRVHV5', 'DCSRVHV9', 'DCSRVHV14', 'DCSRVHV2', 'DCSRVHV6', 'DCSRVHV7', 'DCSRVHV11', 'DCSRVHV15', 'DCSRVHVTP', 'DCSRVHVTSG', 'DCSRVHV4';
$result = New-Object System.Collections.Arraylist;
    $servers = Get-VM -ComputerName $hvList
foreach ($s in $servers)
{
        $state = $s.State;
    
        if ($state -eq 2) {
            $state = "Running";
        } else {
            $state = "Off";
        }
    
        $vm = @{
            "vmid" = $s.Id;
			"name" = $s.Name;
            "state" = $state;
			"hv" = $s.ComputerName;
        };

        $result.Add($vm) | Out-Null
    }
    
    $result | ConvertTo-Json -AsArray -Compress;
