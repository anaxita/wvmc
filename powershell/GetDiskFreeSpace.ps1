param (
    [string]$ip,
    [string]$u,
    [string]$p
)

[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding("utf-8")
$pass = ConvertTo-SecureString -String  $p -AsPlainText -Force
$Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $u, $pass

$result = Invoke-Command -ComputerName $ip -Authentication Negotiate -Credential $Creds -ScriptBlock {
    Get-Volume |
        Where-Object {$_.DriveLetter -ne $null -AND [math]::round($_.Size / 1GB) -ge 1}
}

$result |
    Select-Object @{
        Name='disk_letter'; Expression={$_.DriveLetter}},
        @{Name='space_total'; Expression={[math]::round($_.Size / 1GB, 2)}},
        @{Name='space_free'; Expression={[math]::round($_.SizeRemaining / 1GB, 2)}} |
            ConvertTo-Json -AsArray