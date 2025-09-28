# Banned CLI Windows Installer
# Usage: 
#   From PowerShell: iwr -useb https://github.com/daniel-le97/banned-cli/releases/latest/download/install.ps1 | iex
#   Or download and run: .\install.ps1

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\banned-cli",
    [switch]$Force
)

# Colors for output
$Red = "`e[31m"
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

function Write-Info {
    param([string]$Message)
    Write-Host "${Blue}ℹ${Reset} $Message"
}

function Write-Success {
    param([string]$Message)
    Write-Host "${Green}✅${Reset} $Message"
}

function Write-Warning {
    param([string]$Message)
    Write-Host "${Yellow}⚠${Reset} $Message"
}

function Write-Error {
    param([string]$Message)
    Write-Host "${Red}❌${Reset} $Message"
}

function Write-Step {
    param([string]$Message)
    Write-Host "${Blue}🚀${Reset} $Message"
}

function Test-CommandExists {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        return "x86_64"
    } else {
        return "i386"
    }
}

function Get-LatestVersion {
    try {
        Write-Info "Fetching latest version from GitHub..."
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/daniel-le97/banned-cli/releases/latest" -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to fetch latest version: $($_.Exception.Message)"
        exit 1
    }
}

function Download-File {
    param(
        [string]$Url,
        [string]$OutputPath
    )
    
    try {
        Write-Info "Downloading from: $Url"
        
        # Use System.Net.WebClient for better compatibility
        $webClient = New-Object System.Net.WebClient
        $webClient.Headers.Add("User-Agent", "banned-cli-installer/1.0")
        $webClient.DownloadFile($Url, $OutputPath)
        $webClient.Dispose()
        
        return $true
    }
    catch {
        Write-Error "Download failed: $($_.Exception.Message)"
        return $false
    }
}

function Expand-Archive {
    param(
        [string]$Path,
        [string]$DestinationPath
    )
    
    try {
        if (Test-Path $DestinationPath) {
            Remove-Item -Path $DestinationPath -Recurse -Force
        }
        
        # Use built-in Expand-Archive if available (PowerShell 5+)
        if (Get-Command Expand-Archive -ErrorAction SilentlyContinue) {
            Expand-Archive -Path $Path -DestinationPath $DestinationPath -Force
        }
        else {
            # Fallback for older PowerShell versions
            Add-Type -AssemblyName System.IO.Compression.FileSystem
            [System.IO.Compression.ZipFile]::ExtractToDirectory($Path, $DestinationPath)
        }
        
        return $true
    }
    catch {
        Write-Error "Failed to extract archive: $($_.Exception.Message)"
        return $false
    }
}

function Add-ToPath {
    param([string]$Directory)
    
    # Get current user PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
    
    if ($currentPath -notlike "*$Directory*") {
        Write-Step "Adding $Directory to user PATH..."
        $newPath = "$Directory;$currentPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
        
        # Also update current session PATH
        $env:PATH = "$Directory;$env:PATH"
        
        Write-Success "Added to PATH successfully"
        return $true
    }
    else {
        Write-Info "Directory already in PATH"
        return $false
    }
}

function Initialize-Database {
    param([string]$BannedPath)
    
    Write-Step "Initializing database..."
    
    try {
        # Check if banned.db exists in the installation directory
        $dbPath = Join-Path $InstallDir "banned.db"
        if (Test-Path $dbPath) {
            Write-Info "Database file found at: $dbPath"
            
            # Set environment variable for the CLI to find the database
            [Environment]::SetEnvironmentVariable("BANNED_DB_PATH", $dbPath, [EnvironmentVariableTarget]::User)
            $env:BANNED_DB_PATH = $dbPath
            
            Write-Success "Database initialized successfully"
        }
        else {
            Write-Warning "No database file found. The CLI will create one on first run."
        }
    }
    catch {
        Write-Warning "Could not initialize database: $($_.Exception.Message)"
    }
}

# Main installation process
Write-Host ""
Write-Host "${Blue}🚀 banned CLI Windows Installer${Reset}"
Write-Host "=================================="
Write-Host ""

# Check if running as administrator (optional, but show warning)
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Not running as Administrator. Installing to user directory: $InstallDir"
}

# Determine version to install
if ($Version -eq "latest") {
    $Version = Get-LatestVersion
}

Write-Info "Installing banned CLI version: $Version"

# Determine architecture
$arch = Get-Architecture()
Write-Info "Detected architecture: $arch"

# Create installation directory
if (-not (Test-Path $InstallDir)) {
    Write-Step "Creating installation directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}
elseif (-not $Force) {
    $existing = Get-ChildItem $InstallDir -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Warning "Installation directory exists and contains files."
        $response = Read-Host "Continue anyway? (y/N)"
        if ($response -notmatch '^[yY]') {
            Write-Info "Installation cancelled."
            exit 0
        }
    }
}

# Download the release
$filename = "banned_Windows_${arch}.zip"
$downloadUrl = "https://github.com/daniel-le97/banned-cli/releases/download/$Version/$filename"
$downloadPath = Join-Path $env:TEMP $filename

Write-Step "Downloading banned CLI..."
if (-not (Download-File -Url $downloadUrl -OutputPath $downloadPath)) {
    Write-Error "Failed to download banned CLI"
    exit 1
}

Write-Success "Downloaded successfully"

# Extract the archive
Write-Step "Extracting archive..."
$extractPath = Join-Path $env:TEMP "banned-extract"
if (-not (Expand-Archive -Path $downloadPath -DestinationPath $extractPath)) {
    Write-Error "Failed to extract archive"
    exit 1
}

# Find and move the binary
$binaryPath = Get-ChildItem -Path $extractPath -Name "banned.exe" -Recurse | Select-Object -First 1
if (-not $binaryPath) {
    Write-Error "Could not find banned.exe in the downloaded archive"
    exit 1
}

$sourceBinary = Join-Path $extractPath $binaryPath
$targetBinary = Join-Path $InstallDir "banned.exe"

Write-Step "Installing binary to: $targetBinary"
Copy-Item -Path $sourceBinary -Destination $targetBinary -Force

# Download database file if it exists
Write-Step "Downloading database file..."
$dbUrl = "https://github.com/daniel-le97/banned-cli/releases/download/$Version/banned.db"
$dbPath = Join-Path $InstallDir "banned.db"

if (Download-File -Url $dbUrl -OutputPath $dbPath) {
    Write-Success "Database downloaded successfully"
}
else {
    Write-Warning "Could not download database file (this is optional)"
}

# Add to PATH
$pathAdded = Add-ToPath -Directory $InstallDir

# Initialize database
Initialize-Database -BannedPath $targetBinary

# Clean up temporary files
Write-Step "Cleaning up temporary files..."
Remove-Item -Path $downloadPath -Force -ErrorAction SilentlyContinue
Remove-Item -Path $extractPath -Recurse -Force -ErrorAction SilentlyContinue

# Verify installation
Write-Step "Verifying installation..."
try {
    $version = & $targetBinary --version 2>$null
    Write-Success "Installation completed successfully!"
    Write-Host ""
    Write-Host "${Green}banned CLI $Version installed successfully!${Reset}"
    Write-Host ""
    Write-Info "Installation location: $InstallDir"
    Write-Info "Binary path: $targetBinary"
    
    if ($pathAdded) {
        Write-Warning "PATH updated. You may need to restart your terminal/PowerShell session."
        Write-Info "Or run: `$env:PATH = `"$InstallDir;`$env:PATH`""
    }
    
    Write-Host ""
    Write-Host "${Blue}Usage:${Reset}"
    Write-Host "  banned --help          # Show help"
    Write-Host "  banned version         # Show version"
    Write-Host "  banned download <url>  # Download a video"
    Write-Host ""
    
    # Test if banned is now in PATH
    if (Test-CommandExists "banned") {
        Write-Success "You can now run 'banned' from anywhere!"
    }
    else {
        Write-Info "Run the following to use banned in this session:"
        Write-Host "  $targetBinary"
        Write-Host "Or restart your terminal to use 'banned' directly."
    }
}
catch {
    Write-Error "Installation verification failed: $($_.Exception.Message)"
    Write-Info "You can try running: $targetBinary --version"
    exit 1
}

Write-Host ""
Write-Success "Installation complete! 🎉"