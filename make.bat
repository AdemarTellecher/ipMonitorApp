@echo off
REM Build do IP Monitor App (Wails v3 - Executável Único, Leve e Portátil)
set APP_NAME=ipMonitorApp
set SRC=.

REM Garante que o GCC (WinLibs MinGW-w64) e o Go bin estejam no PATH
set "PATH=%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;%USERPROFILE%\go\bin;%PATH%"
set CGO_ENABLED=1

echo Compilando executavel unico e enxuto (Wails v3 + Assets Embutidos)...
go build -ldflags="-s -w -H=windowsgui" -o %APP_NAME%.exe %SRC%
if %ERRORLEVEL%==0 (
    echo.
    echo ========================================================
    echo Build concluido com sucesso: %APP_NAME%.exe
    echo - Binario unico e autocontido - Frontend embutido via go:embed
    echo - Sem DLLs externas ou pastas adicionais necessarias
    echo - SQLite portatil: cria/usa ipmonitor.db na pasta de execucao
    echo ========================================================
) else (
    echo Erro na compilacao!
)
