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
    export CGO_ENABLED := 0

    RM_CMD = cmd /C if exist $(BINARY) del /Q /F $(BINARY)
    RUN_CMD = .\$(BINARY)
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := macOS (Darwin)
        APP_BUNDLE := "IP Monitor.app"
    else
        DETECTED_OS := Linux
    endif

    BINARY := $(APP_NAME)
    LDFLAGS := -s -w
    export CGO_ENABLED := 1

    RM_CMD = rm -rf $(BINARY) $(APP_BUNDLE)
    RUN_CMD = ./$(BINARY)
endif

# ------------------------------------------------------------------------------
# Alvos
# ------------------------------------------------------------------------------
.PHONY: all build run test clean info bundle-mac

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
ifeq ($(DETECTED_OS),macOS (Darwin))
	@echo Empacotando $(APP_BUNDLE)...
	@mkdir -p $(APP_BUNDLE)/Contents/MacOS $(APP_BUNDLE)/Contents/Resources
	@cp $(BINARY) $(APP_BUNDLE)/Contents/MacOS/$(BINARY)
	@cp build/darwin/Info.plist $(APP_BUNDLE)/Contents/Info.plist
	@if [ -f internal/assets/icons/app-icon.png ]; then \
		mkdir -p AppIcon.iconset; \
		sips -z 16 16     internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_16x16.png >/dev/null 2>&1 || true; \
		sips -z 32 32     internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_16x16@2x.png >/dev/null 2>&1 || true; \
		sips -z 32 32     internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_32x32.png >/dev/null 2>&1 || true; \
		sips -z 64 64     internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_32x32@2x.png >/dev/null 2>&1 || true; \
		sips -z 128 128   internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_128x128.png >/dev/null 2>&1 || true; \
		sips -z 256 256   internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_128x128@2x.png >/dev/null 2>&1 || true; \
		sips -z 256 256   internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_256x256.png >/dev/null 2>&1 || true; \
		sips -z 512 512   internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_256x256@2x.png >/dev/null 2>&1 || true; \
		sips -z 512 512   internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_512x512.png >/dev/null 2>&1 || true; \
		sips -z 1024 1024 internal/assets/icons/app-icon.png --out AppIcon.iconset/icon_512x512@2x.png >/dev/null 2>&1 || true; \
		iconutil -c icns AppIcon.iconset -o $(APP_BUNDLE)/Contents/Resources/AppIcon.icns >/dev/null 2>&1 || true; \
		rm -rf AppIcon.iconset; \
	fi
	@echo Assinando $(APP_BUNDLE) com ad-hoc codesign...
	@codesign --force --deep --sign - $(APP_BUNDLE) >/dev/null 2>&1 || true
	@echo Pacote $(APP_BUNDLE) criado e assinado com sucesso!
endif
	@echo Build concluido com sucesso: $(BINARY)

run: build
	@echo Iniciando $(BINARY)...
ifeq ($(DETECTED_OS),macOS (Darwin))
	open $(APP_BUNDLE)
else
	$(RUN_CMD)
endif

test:
	@echo Executando testes unitarios...
	go test -v ./...

clean:
	@echo Removendo arquivos de build...
	$(RM_CMD)
	@echo Limpeza concluida.
