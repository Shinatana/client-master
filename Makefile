LOCAL_BIN:=$(CURDIR)/bin

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

lint:
	$(LOCAL_BIN)/golangci-lint run --config=./golangci.yml

lint-fix:
	$(LOCAL_BIN)/golangci-lint run --config=./golangci.yml --fix

test:
	go test -v ./...
