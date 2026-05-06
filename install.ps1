<#
.SYNOPSIS
Installer script for chesshell CLI tool on Windows.

.DESCRIPTION
This script automatically downloads the latest release of chesshell-windows-amd64.exe
from GitHub, places it in the user's LocalAppData folder, and adds that folder
to the user's PATH environment variable.
#>

$ErrorActionPreference = "Stop"

$Repo = "Killiivalavan/chesshell-cli"
$BinName = "chesshell.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "chesshell\bin"
$BinPath = Join-Path $InstallDir $BinName
$AssetPattern = "chesshell-windows-x86_64*" # This will match our generated asset. Wait, we built chesshell-windows-amd64.exe

# Correct asset pattern
$AssetPattern = "chesshell-windows-amd64.exe"

Write-Host "Fetching latest release information..."
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
$ReleaseData = Invoke-RestMethod -Uri $ApiUrl

$DownloadUrl = $null
foreach ($asset in $ReleaseData.assets) {
    if ($asset.name -eq $AssetPattern) {
        $DownloadUrl = $asset.browser_download_url
        break
    }
}

if (-not $DownloadUrl) {
    Write-Error "Could not find release asset for Windows amd64."
}

Write-Host "Downloading chesshell to $InstallDir..."
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinPath

Write-Host "Installation successful!"

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to your PATH..."
    $NewPath = $UserPath + ";$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    
    Write-Host ""
    Write-Host "==========================================================="
    Write-Host "chesshell has been installed to $InstallDir and added to your PATH."
    Write-Host "You must RESTART YOUR TERMINAL (or open a new tab) for the PATH changes to take effect."
    Write-Host "Then, you can run: chesshell help"
    Write-Host "==========================================================="
} else {
    Write-Host ""
    Write-Host "==========================================================="
    Write-Host "chesshell has been updated at $InstallDir."
    Write-Host "Run: chesshell help"
    Write-Host "==========================================================="
}
