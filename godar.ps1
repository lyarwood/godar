#Requires -RunAsAdministrator

param (
    [Parameter(Mandatory=$true, Position=0)]
    [ValidateSet("install", "uninstall", "start", "stop", "status")]
    [string]$Command,

    [string]$ServiceName = "godar",
    [string]$DisplayName = "Godar Aircraft Monitor",
    [string]$Description = "Monitors overflying aircraft from a Virtual Radar Server.",
    [string]$ExecutablePath = ""
)

function Test-ServiceExists {
    param ([string]$Name)
    $service = Get-Service -Name $Name -ErrorAction SilentlyContinue
    if ($null -ne $service) {
        return $true
    }
    return $false
}

switch ($Command) {
    "install" {
        if (Test-ServiceExists -Name $ServiceName) {
            Write-Error "Service '$ServiceName' already exists."
            exit 1
        }

        if (-not $ExecutablePath) {
            # Default to the path of the script if not provided
            $ExecutablePath = Join-Path -Path $PSScriptRoot -ChildPath "godar.exe"
        }

        if (-not (Test-Path -Path $ExecutablePath)) {
            Write-Error "Executable not found at '$ExecutablePath'. Please provide the correct path."
            exit 1
        }

        Write-Host "Installing service '$ServiceName'..."
        New-Service -Name $ServiceName -BinaryPathName "$ExecutablePath monitor" -DisplayName $DisplayName -Description $Description -StartupType Automatic
        Write-Host "Service '$ServiceName' installed successfully."
    }
    "uninstall" {
        if (-not (Test-ServiceExists -Name $ServiceName)) {
            Write-Error "Service '$ServiceName' does not exist."
            exit 1
        }

        Write-Host "Uninstalling service '$ServiceName'..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        # Using sc.exe delete as Remove-Service is not available on older PowerShell versions
        sc.exe delete $ServiceName
        Write-Host "Service '$ServiceName' uninstalled successfully."
    }
    "start" {
        if (-not (Test-ServiceExists -Name $ServiceName)) {
            Write-Error "Service '$ServiceName' does not exist."
            exit 1
        }

        Write-Host "Starting service '$ServiceName'..."
        Start-Service -Name $ServiceName
        Write-Host "Service '$ServiceName' started."
    }
    "stop" {
        if (-not (Test-ServiceExists -Name $ServiceName)) {
            Write-Error "Service '$ServiceName' does not exist."
            exit 1
        }

        Write-Host "Stopping service '$ServiceName'..."
        Stop-Service -Name $ServiceName
        Write-Host "Service '$ServiceName' stopped."
    }
    "status" {
        if (-not (Test-ServiceExists -Name $ServiceName)) {
            Write-Error "Service '$ServiceName' does not exist."
            exit 1
        }

        Get-Service -Name $ServiceName | Select-Object -Property Name, DisplayName, Status
    }
}
