APP_NAME := grpc-rest-proxy
VERSION  := $(shell git describe --tags 2> /dev/null || echo dev)
COMMIT   := $(shell git rev-parse HEAD)
BUILD    := $(shell date +%FT%T%z)
PATH     := $(GOPATH)/bin:/usr/local/bin:/usr/local/go/bin:$(PATH)

VERSION_FILE := main
LDFLAGS=-ldflags "-X=$(VERSION_FILE).Version=$(VERSION) -X=$(VERSION_FILE).Commit=$(COMMIT) -X=$(VERSION_FILE).Build=$(BUILD)"

all: vulncheck build test race cover
default: build
build: fmt lint build-only
test: fmt lint test-only

fmt:
	go fmt ./...

lint:
	@golangci-lint --version > /dev/null 2>&1 || { echo >&2 "golangci-lint not installed. See: https://golangci-lint.run/usage/install/#linux-and-windows."; exit 1; }
	golangci-lint run

vulncheck:
	govulncheck ./...

build-only: build-service build-example-grpc-server

build-service:
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o $(APP_NAME) ./cmd/service/

build-example-grpc-server:
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o example-grpc-server ./cmd/examples/grpcserver	

test-only:
	go test ./...

race:
	go test -race ./...

cover:
	go test -cover ./...

coverprofile:
	go test -coverprofile=coverage.out  ./...
	go tool cover -html=coverage.out

clean:
	rm -f $(APP_NAME)

docker:
	go mod vendor
	docker build . -t grpc-rest-proxy-test

generate:
	buf generate

.PHONY: all default build test fmt lint build-only test-only race cover coverprofile clean