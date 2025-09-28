@echo off
setlocal enabledelayedexpansion

:: Banned CLI Windows Installer (Batch Version)
:: This is a fallback for systems where PowerShell execution is restricted

echo.
echo 🚀 banned CLI Windows Installer (Batch Version)
echo ==========================================
echo.

:: Check if PowerShell is available and try the main installer
where powershell >nul 2>&1
if !errorlevel! equ 0 (
    echo ℹ PowerShell detected. Attempting to use main installer...
    powershell -ExecutionPolicy Bypass -Command "& { iwr -useb https://github.com/daniel-le97/banned-cli/releases/latest/download/install.ps1 | iex }"
    if !errorlevel! equ 0 (
        echo.
        echo ✅ Installation completed successfully!
        pause
        exit /b 0
    )
    echo.
    echo ⚠ PowerShell installer failed. Falling back to manual method...
)

:: Manual installation fallback
echo.
echo 📋 Manual Installation Instructions
echo ==================================
echo.
echo 1. Go to: https://github.com/daniel-le97/banned-cli/releases/latest
echo 2. Download: banned_Windows_x86_64.zip (64-bit) or banned_Windows_i386.zip (32-bit)
echo 3. Extract the zip file
echo 4. Move banned.exe to a folder in your PATH (e.g., C:\Windows\System32)
echo 5. Or add the folder containing banned.exe to your PATH environment variable
echo.
echo 🌐 For detailed instructions, visit:
echo    https://github.com/daniel-le97/banned-cli#installation
echo.

pause