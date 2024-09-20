.PHONY: run build

run:
	go run main.go

build:
	CGO_ENABLED=0 $(GOOS) go build

docker:
	docker build -t goh .
	docker run --rm -v $(shell pwd):/out goh cp /app/goh /out/
