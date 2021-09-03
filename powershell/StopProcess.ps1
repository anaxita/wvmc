param (
    [string]$ip,
    [string]$u,
    [string]$p,
    [int]$id
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
    Stop-Process -Id $Using:id -Force
} -ErrorAction Stop