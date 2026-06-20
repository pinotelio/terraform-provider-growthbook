GOBIN ?= $(shell go env GOPATH)/bin
VERSION ?= dev

default: build

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./...

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)" .

.PHONY: test
test:
	go test ./... -count=1

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -count=1 -timeout 120m

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docs
docs:
	go generate ./...

.PHONY: tidy
tidy:
	go mod tidy
