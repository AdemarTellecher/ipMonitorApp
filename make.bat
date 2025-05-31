@echo off
REM Build do IP Monitor App para Windows
set APP_NAME=ipMonitorApp
set SRC=cmd/main.go
set ICON=internal/assets/icons/ip-monitor-icon.jpg

echo Compilando para Windows...
go build -ldflags="-H=windowsgui" -o %APP_NAME%-win.exe %SRC%
if %ERRORLEVEL%==0 (
    echo Build concluido: %APP_NAME%-win.exe
) else (
    echo Erro na compilacao!
)
