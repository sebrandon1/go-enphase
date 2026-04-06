APP_NAME=go-enphase
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/sebrandon1/go-enphase/cmd.Version=$(VERSION)"

vet:
	go vet ./...

build:
	go build $(LDFLAGS) -o $(APP_NAME)

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -f $(APP_NAME)

install:
	go install $(LDFLAGS) ./...

.PHONY: vet build lint test clean install
