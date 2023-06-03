#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# TODO: Right now putting here there is no way to specifiy where we have the configs
build:
	go build -o ./bin/local-ic ./src/local-ic

install:
	go install -mod=readonly ./src/local-ic

run:
	./bin/local-ic

.PHONY: build install run