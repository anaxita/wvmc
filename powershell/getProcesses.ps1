param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

$prcList = Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
Get-Process -IncludeUserName | Select-Object -Property UserName, StartTime, PagedMemorySize, Id, Name  | Sort-Object -Property WorkingSet -Descending
}

$prcListRezult = $prcList | ForEach-Object {
    $userName = $_.UserName -split "\\"

    if ($Null -ne $userName[1]) {
        $cpuLoad = 0

        if ($Null -ne $_.StartTime) {
            $totalUpTime = (New-TimeSpan -Start $_.StartTime).TotalSeconds # for get correct cpu % load
            $cpuLoad = [Math]::Round( ($_.CPU * 100 / $totalUpTime))
        }

        $memoryUsed = [Math]::Round($_.PagedMemorySize / 1MB)  
        
        [PSCustomObject]@{
            "id"        = [int]$_.Id
            "name"      = [string]$_.Name 
            "user_name" = [string]$userName[1]
            "cpu_load"  = [int]$cpuLoad
            "memory"    = [int]$memoryUsed
        }

    }
}

$result = @{}
$prcListRezult | ForEach-Object {
    $uname = $_.user_name
    if ($Null -ne $result.$uname) {
        $result.$uname += $_
    }
    else {   
        $result.$uname = @()
        $result.$uname += $_
    }
}

Write-Output $result | ConvertTo-Json -Compress