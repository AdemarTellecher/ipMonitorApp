# Makefile para build multiplataforma do IP Monitor App

APP_NAME=ipMonitorApp
SRC=cmd/main.go
ICON=internal/assets/icons/ip-monitor-icon.jpg

.PHONY: all linux macos android ios clean

all: linux macos android ios

linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(SRC)

macos:
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-macos $(SRC)

android:
	fyne package -os android -icon $(ICON) -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

ios:
	fyne package -os ios -icon $(ICON) -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

clean:
	rm -f $(APP_NAME)-linux $(APP_NAME)-macos $(APP_NAME)-win.exe
