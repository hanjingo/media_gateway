export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

VERSION = 0.1.0
TARGET = gate

BIN = ./bin
MAIN = ./
MAIN_SRC = ./main.go

ARM = ./bin/arm
ARM64_TARGET = $(TARGET).arm64.v$(VERSION)

LINUX = ./bin/linux
LINUX64_TARGET = $(TARGET).amd64.v$(VERSION)

WIN = ./bin/windows
WIN64_TARGET = $(TARGET).win64.v$(VERSION).exe

INSTALL = /usr/local/bin
DATA = $(INSTALL)/media_gate

.PHONEY: all arm64 linux64 win64

all: linux64

arm64: 
	mkdir -p $(ARM)
	cp $(MAIN)/conf.json $(ARM)
	cp -rf $(MAIN)/statics $(ARM)
	env GOOS=linux GOARCH=arm64 GOARM=7 go build -o $(ARM)/$(ARM64_TARGET) $(MAIN_SRC)

linux64: 
	mkdir -p $(LINUX)
	cp $(MAIN)/conf.json $(LINUX)
	cp -rf $(MAIN)/statics $(LINUX)
	env GOOS=linux GOARCH=amd64 go build -o $(LINUX)/$(LINUX64_TARGET) $(MAIN_SRC)

win64:
	mkdir -p $(WIN)
	cp $(MAIN)/conf.json $(WIN)
	cp -rf $(MAIN)/statics $(WIN)
	env GOOS=windows GOARCH=amd64 go build -o $(WIN)/$(WIN64_TARGET) $(MAIN_SRC)

install:
	mkdir -p $(DATA)
	cp -rf $(MAIN)/statics $(DATA)
	cp $(MAIN)/conf.json $(INSTALL)

	cp $(LINUX)/$(LINUX64_TARGET) $(DATA)
	ln -snf $(DATA)/$(LINUX64_TARGET) $(INSTALL)/gate

clean:
	rm -rf $(ARM)
	rm -rf $(LINUX)
	rm -rf $(WIN)