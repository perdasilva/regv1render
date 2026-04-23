include .bingo/Variables.mk

BINARY  := rv1
BIN_DIR := bin
CMD_DIR := ./cmd/rv1

.PHONY: build test lint fmt vet tidy verify clean

build:
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

test:
	go test ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

fmt: $(GOIMPORTS)
	gofmt -w .
	$(GOIMPORTS) -w -local github.com/perdasilva/rv1 .

vet:
	go vet ./...

tidy:
	go mod tidy

verify: fmt vet lint test

clean:
	rm -rf $(BIN_DIR)
