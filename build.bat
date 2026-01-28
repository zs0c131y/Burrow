@echo off
REM Burrow Build Script for Windows

echo ====================================
echo Building Burrow...
echo ====================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    exit /b 1
)

echo Go version:
go version
echo.

REM Clean previous builds
if exist wm.exe (
    echo Removing previous build...
    del wm.exe
)

REM Download dependencies
echo Downloading dependencies...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to download dependencies
    exit /b 1
)

REM Build the application
echo Building Burrow...
go build -ldflags="-s -w" -o wm.exe
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Build failed
    exit /b 1
)

echo.
echo ====================================
echo Build successful!
echo ====================================
echo.
echo Binary: wm.exe
echo.

REM Show file size
for %%A in (wm.exe) do (
    echo Size: %%~zA bytes
)

echo.
echo To install globally, copy wm.exe to a directory in your PATH
echo Example: copy wm.exe C:\Windows\System32\
echo.
echo To run: wm --help
echo.

pause
