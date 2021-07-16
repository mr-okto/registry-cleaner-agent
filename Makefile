.PHONY: build
build:
	go build -o build/agent -v ./cmd

.DEFAULT_GOAL := build