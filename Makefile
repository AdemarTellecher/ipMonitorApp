# Makefile para build multiplataforma do IP Monitor App

APP_NAME=ipMonitorApp
SRC=cmd/main.go
ICON=internal/assets/icons/ip-monitor-icon.jpg
FYNE_DLLS=libEGL.dll libGLESv2.dll dwrite.dll libpng16-16.dll zlib1.dll
FYNE_DLLS_PATH=$(shell go list -f '{{.Dir}}' -m fyne.io/fyne/v2)/internal/driver/windows/dlls

.PHONY: all linux macos android ios windows clean

all: linux macos android ios windows

linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(SRC)

macos:
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-macos $(SRC)

android:
	fyne package -os android -icon $(ICON) -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

ios:
	fyne package -os ios -icon $(ICON) -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o $(APP_NAME)-win.exe $(SRC)
	@for dll in $(FYNE_DLLS); do \
	  cp $(FYNE_DLLS_PATH)/$$dll . 2>/dev/null || true; \
	done

clean:
	rm -f $(APP_NAME)-linux $(APP_NAME)-macos $(APP_NAME)-win.exe *.dll
