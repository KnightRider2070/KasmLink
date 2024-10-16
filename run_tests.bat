@echo off
setlocal

REM Set environment variables for cross-compilation
set GOARCH=amd64
set GOOS=windows

REM Allow specifying a project directory as an argument or use the current directory as default
set "PROJECT_DIR=%~1"
if "%PROJECT_DIR%"=="" set "PROJECT_DIR=."

REM Navigate to the project directory
echo Navigating to the project directory: %PROJECT_DIR%
cd "%PROJECT_DIR%" || (
    echo Failed to navigate to the project directory
    exit /b 1
)

REM Clean the Go cache and temporary files
echo Cleaning Go cache...
go clean -cache -modcache -i -r
if %errorlevel% neq 0 (
    echo Failed to clean Go cache
    exit /b %errorlevel%
)

REM Build the project
echo Building the project...
go build
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
) else (
    echo Build succeeded!
)

REM Run the tests in verbose mode
echo Running tests...
go test -v ./...
if %errorlevel% neq 0 (
    echo Tests failed!
    exit /b %errorlevel%
) else (
    echo All tests passed successfully!
)

REM Clean up build artifacts (optional)
echo Cleaning build artifacts...
go clean
if %errorlevel% neq 0 (
    echo Failed to clean build artifacts
    exit /b %errorlevel%
) else (
    echo Cleaned up build artifacts successfully!
)

REM Done
echo Script executed successfully!

REM Pause the command window to see the output
pause
