@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
set "OUTPUT_DIR=%SCRIPT_DIR%icoo_runtime"
set "OUTPUT_PATH=%OUTPUT_DIR%\icoo.exe"
set "CONFIG_SOURCE=%SCRIPT_DIR%config.toml.example"
set "CONFIG_TARGET=%OUTPUT_DIR%\config.toml"
set "PACKAGE_PATH=./icoo_runtime/cmd/assistant"

if not exist "%SCRIPT_DIR%icoo_runtime\cmd\assistant" set "PACKAGE_PATH=./cmd/assistant"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

pushd "%SCRIPT_DIR%"
go build -o "%OUTPUT_PATH%" %PACKAGE_PATH%
set "EXIT_CODE=%ERRORLEVEL%"
if exist "%CONFIG_SOURCE%" copy /Y "%CONFIG_SOURCE%" "%CONFIG_TARGET%" >nul
popd

if errorlevel 1 exit /b %EXIT_CODE%
echo Built %OUTPUT_PATH% from %PACKAGE_PATH%
if exist "%CONFIG_TARGET%" echo Copied %CONFIG_TARGET% from %CONFIG_SOURCE%

@REM upx
upx "%OUTPUT_PATH%"

endlocal
