param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

$result = Invoke-Command -ComputerName $ip -Credential $Creds -ScriptBlock {
        Get-WmiObject -Class win32_service
    }

$result |
    ForEach-Object {
        [PSCustomObject]@{
            'name'         = $_.Name
            'display_name' = $_.DisplayName
            'status'       = [string]$_.State
            'user'         = $_.StartName
        }
    } |
    ConvertTo-Json -Compress