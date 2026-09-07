# Makefile para build do IP Monitor App (Wails v3)

APP_NAME=ipMonitorApp
SRC=.

.PHONY: all build clean test run

all: build

build:
	go build -ldflags="-s -w -H=windowsgui" -o $(APP_NAME).exe $(SRC)

test:
	go test -v ./...

clean:
	rm -f $(APP_NAME).exe

run: build
	./$(APP_NAME).exe
