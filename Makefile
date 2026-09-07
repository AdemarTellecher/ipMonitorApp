# ==============================================================================
# Makefile para IP Monitor App (Wails v3)
# Executável Único, Leve e Portátil com Auto-detecção de Sistema Operacional
# ==============================================================================

APP_NAME := ipMonitorApp
SRC := .

# ------------------------------------------------------------------------------
# Auto-detecção de Sistema Operacional
# ------------------------------------------------------------------------------
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY := $(APP_NAME).exe
    LDFLAGS := -s -w -H=windowsgui
    export CGO_ENABLED := 1

    # Inclui o MinGW GCC (WinLibs) e o binário do Go no PATH no ambiente Windows
    WINLIBS_BIN := $(LOCALAPPDATA)/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin
    export PATH := $(WINLIBS_BIN):$(USERPROFILE)/go/bin:$(PATH)

    RM_CMD = cmd /C if exist $(BINARY) del /Q /F $(BINARY)
    RUN_CMD = .\$(BINARY)
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := macOS (Darwin)
    else
        DETECTED_OS := Linux
    endif

    BINARY := $(APP_NAME)
    LDFLAGS := -s -w
    export CGO_ENABLED := 1

    RM_CMD = rm -f $(BINARY)
    RUN_CMD = ./$(BINARY)
endif

# ------------------------------------------------------------------------------
# Alvos
# ------------------------------------------------------------------------------
.PHONY: all build run test clean info

all: build

info:
	@echo ========================================================
	@echo  Sistema Operacional detectado: $(DETECTED_OS)
	@echo  Binario de saida:              $(BINARY)
	@echo  Flags de linker (ldflags):     "$(LDFLAGS)"
	@echo ========================================================

build: info
	@echo Compilando executavel unico e portatil...
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(SRC)
	@echo Build concluido com sucesso: $(BINARY)

run: build
	@echo Iniciando $(BINARY)...
	$(RUN_CMD)

test:
	@echo Executando testes unitarios...
	go test -v ./...

clean:
	@echo Removendo binario $(BINARY)...
	$(RM_CMD)
	@echo Limpeza concluida.
