param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
    $prcListRezult = Get-Process | Where-Object {$_.SessionId -ne 0} | Select-Object UserName, StartTime, PagedMemorySize, Id, Name, SI, SessionId | ForEach-Object {    
                $cpuLoad = 0
                $sessionID = [int]$_.SI
    
                if ($Null -ne $_.StartTime) {
                    $totalUpTime = (New-TimeSpan -Start $_.StartTime).TotalSeconds # for get correct cpu % load
                    $cpuLoad = [Math]::Round( ($_.CPU * 100 / $totalUpTime))
                }
    
                $memoryUsed = [Math]::Round($_.PagedMemorySize / 1MB)  
            
                [PSCustomObject]@{
                    "session_id" = $sessionID
                    "id"         = [int]$_.Id
                    "name"       = [string]$_.Name 
                    "cpu_load"   = [int]$cpuLoad
                    "memory"     = [int]$memoryUsed
                }
    }
    
    $rdpSessions = Get-TSSession |
        Where-Object { $_.UserName -ne '' -and $_.SessionId -ne 0 } |
            Select-Object -Property UserName, State, SessionId |
                ForEach-Object {
                    [PSCustomObject]@{
                        "session_id" = [int]$_.SessionId
                        "user_name"  = [string]$_.UserName 
                        "state"      = [string]$_.State
                        "processes"  = @()        
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

    if ($Null -eq $rdpSessions) {
        Write-Output @()
    } else {
        Write-Output $rdpSessions
    }

} | ConvertTo-Json -AsArray -Depth 3 -Compress
