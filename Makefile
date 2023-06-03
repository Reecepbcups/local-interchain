#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# MacOS / Linux only?
SHELL := /bin/bash
CWD := $(shell cd -P -- '$(shell dirname -- "$0")' && pwd -P)

ldflags = -X main.InstallDirectory=$(CWD) -X main.Commit=$(COMMIT) -X main.Version=$(VERSION)
ldflags := $(strip $(ldflags))
BUILD_FLAGS := -ldflags '$(ldflags)'

build:
	go build $(BUILD_FLAGS) -o ./bin/local-ic ./src/local-ic 

install:
	go install $(BUILD_FLAGS) -mod=readonly ./src/local-ic

run:
	./bin/local-ic

.PHONY: build install run