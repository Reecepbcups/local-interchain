# TODO: Right now putting here there is no way to specifiy where we have the configs
build:
	go build -o ./bin/local-ic ./src

run:
	./bin/local-ic

.PHONY: build run