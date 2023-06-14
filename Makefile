#!/usr/bin/make -f

CWD := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

ldflags = -X main.MakeFileInstallDirectory=$(CWD)
ldflags := $(strip $(ldflags))
BUILD_FLAGS := -ldflags '$(ldflags)'

build:
	go build $(BUILD_FLAGS) -o ./bin/local-ic ./cmd/local-ic 

run:
	./bin/local-ic

install:
	go install $(BUILD_FLAGS) ./cmd/local-ic ./interchain

.PHONY: build install run