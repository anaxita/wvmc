# param (
#     [string]$ip
#     [string]$u
#     [string]$p
# )


# $pass = ConvertTo-SecureString -String '#TerraCat82Like!' -AsPlainText -Force
# $user = 'KMSservice1'
# $Creds = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $user, $pass

# Invoke-Command -ComputerName 172.16.0.110  -Credential $Creds -ScriptBlock {
    $services = Get-Service | Select-Object Name, DisplayName, Status, UserName -First 30 | ForEach-Object {

        [PSCustomObject]@{
            'name' = $_.Name
            'display_name' = $_.DisplayName
            'status' = [string]$_.Status
            'user' = $_.UserName
        }
    }

    $services | ConvertTo-Json -AsArray -Compress
# } -ErrorAction 'Stop'
