# Makefile para build multiplataforma do IP Monitor App

APP_NAME=ipMonitorApp
SRC=main.go

.PHONY: all windows linux macos clean

all: windows linux macos

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o $(APP_NAME)-win.exe $(SRC)

linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(SRC)

macos:
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-macos $(SRC)

clean:
	rm -f $(APP_NAME)-win.exe $(APP_NAME)-linux $(APP_NAME)-macos
