#!/usr/bin/make -f


 # Untested for Windows
ifeq ($(OS),Windows_NT)
    # Run commands specific to Windows    
    CWD := $(shell cd)
else
    # Run commands specific to Unix        
    CWD := $(shell cd -P -- '$(shell dirname -- "$0")' && pwd -P)
endif

ldflags = -X main.InstallDirectory=$(CWD)
ldflags := $(strip $(ldflags))
BUILD_FLAGS := -ldflags '$(ldflags)'

build:
	go build $(BUILD_FLAGS) -o ./bin/local-ic ./src/local-ic 

install:
	go install $(BUILD_FLAGS) -mod=readonly ./src/local-ic

run:
	./bin/local-ic

.PHONY: build install run