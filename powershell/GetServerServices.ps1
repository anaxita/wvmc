param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

$result = Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
        Get-Service
    }

$result |
    ForEach-Object {
        [PSCustomObject]@{
            'name'         = $_.Name
            'display_name' = $_.DisplayName
            'status'       = [string]$_.Status
            'user'         = $_.UserName
        }
    } |
    ConvertTo-Json -Compress