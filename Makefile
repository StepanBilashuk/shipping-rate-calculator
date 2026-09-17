.PHONY: build test cover vet fmt check run clean

build:
	go build -o bin/shipping-calculator ./cmd/shipping-calculator

test:
	go test ./...

cover:
	go test ./... -cover

vet:
	go vet ./...

fmt:
	gofmt -l -w .

check: vet
	gofmt -l . | tee /dev/stderr | (! read)
	go test ./... -race

run:
	go run ./cmd/shipping-calculator $(ARGS)

clean:
	rm -rf bin
