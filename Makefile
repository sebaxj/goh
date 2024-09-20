.PHONY: run build

OS ?= "darwin"

run:
	go run main.go

build:
	CGO_ENABLED=0 GOOS=$(OS) go build

docker:
	docker build -t goh .
	docker run --rm -v $(shell pwd):/out goh cp /app/goh /out/
