APP_NAME=go-enphase

vet:
	go vet ./...

build:
	go build -o $(APP_NAME)

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -f $(APP_NAME)

install:
	go install ./...

.PHONY: vet build lint test clean install
