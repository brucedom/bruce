APP := bruce
APP_ENTRY := cmd/main.go
SHELL := /bin/bash
VER := source
ifndef VER
$(error BUILD_VER not set: Run make this way: `make VER=1.0.31`)
endif
ROOTPATH := $(shell echo ${PWD})

.PHONY: all clean setup build deploy run

all: build

clean:
	rm -rf ${ROOTPATH}/.build/
	rm -rf ${ROOTPATH}/vendor/

setup:
	mkdir -p ${ROOTPATH}/.build/bin

package: build zipit

deploy: build-local push

build: clean setup build-local

build-local:
	@go version
	@go get ./cmd/...
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VER}" -o ${ROOTPATH}/.build/bin/$(APP) ${APP_ENTRY}

zipit:
	cd .build/ && zip -r ${APP}-${VER}-${UNAME}.zip ./*
	@echo "package ready under: .build/"
