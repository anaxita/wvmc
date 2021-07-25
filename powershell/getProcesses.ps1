param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

$prcList = Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
    Get-Process -IncludeUserName | Select-Object -Property UserName, StartTime, PagedMemorySize, Id, Name, SessionId  | Sort-Object -Property WorkingSet -Descending
}

$prcListRezult = $prcList | ForEach-Object {
    $userName = $_.UserName -split "\\"

    if ($Null -ne $userName[1]) {
        if ($_.SessionId -ne 0) {
        $cpuLoad = 0
        $sessionID = [int]$_.SessionId

        if ($Null -ne $_.StartTime) {
            $totalUpTime = (New-TimeSpan -Start $_.StartTime).TotalSeconds # for get correct cpu % load
            $cpuLoad = [Math]::Round( ($_.CPU * 100 / $totalUpTime))
        }

        $memoryUsed = [Math]::Round($_.PagedMemorySize / 1MB)  
        
        [PSCustomObject]@{
                "session_id" = $sessionID
                "id"         = [int]$_.Id
                "name"       = [string]$_.Name 
                "user_name"  = [string]$userName[1]
                "cpu_load"   = [int]$cpuLoad
                "memory"     = [int]$memoryUsed
            }
        
        }
    }
}


$rdpSessions = Get-TSSession | Where-Object { $_.UserName -ne '' } | Select-Object -Property UserName, State, SessionId | ForEach-Object {
    [PSCustomObject]@{
            "session_id" = [int]$_.SessionId
            "user_name" = [string]$_.UserName 
            "state"     = [string]$_.State
            "processes" = @()
        
    }
}

$rdpSessions | ForEach-Object {
    $session = $_

    $prcListRezult | ForEach-Object {
        $prc = $_

        if ($prc.session_id -eq $session.session_id) {
            $session.processes += $prc
        }
    }
}

Write-Output $rdpSessions | ConvertTo-Json -Depth 3 -Compress