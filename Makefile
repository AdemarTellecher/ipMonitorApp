# Makefile para build multiplataforma do IP Monitor App

APP_NAME=ipMonitorApp
SRC=main.go

.PHONY: all linux macos android ios clean

all: linux macos android ios

linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(SRC)

macos:
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-macos $(SRC)

android:
	fyne package -os android -icon Icon.png -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

ios:
	fyne package -os ios -icon Icon.png -name "$(APP_NAME)" -appID com.example.$(APP_NAME)

clean:
	rm -f $(APP_NAME)-linux $(APP_NAME)-macos
