param (
    [string]$ip,
    [string]$u,
    [string]$p,
    [string]$name
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
    Restart-Service -Name $name -Force
}