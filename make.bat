@echo off
REM Build do IP Monitor App para Windows
set APP_NAME=ipMonitorApp
set SRC=cmd/main.go
set ICON=internal/assets/icons/ip-monitor-icon.jpg

REM Compila o executável principal
REM O build padrão do Go não gera as DLLs do Fyne, apenas o binário

echo Compilando para Windows...
go build -ldflags="-H=windowsgui" -o %APP_NAME%-win.exe %SRC%
if %ERRORLEVEL%==0 (
    echo Build concluido: %APP_NAME%-win.exe
    REM Copia as DLLs do Fyne (se existirem em uma pasta padrao ou assets)
    if exist %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libEGL.dll copy %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libEGL.dll .
    if exist %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libGLESv2.dll copy %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libGLESv2.dll .
    if exist %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\dwrite.dll copy %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\dwrite.dll .
    if exist %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libpng16-16.dll copy %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\libpng16-16.dll .
    if exist %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\zlib1.dll copy %USERPROFILE%\go\pkg\mod\fyne.io\fyne\v2@*\internal\driver\windows\dlls\zlib1.dll .
    echo (Se necessário, copie manualmente as DLLs do Fyne para a pasta do executável)
) else (
    echo Erro na compilacao!
)
