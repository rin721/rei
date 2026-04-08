GO ?= go

.PHONY: list test vet quality fmt run-server run-server-dry run-initdb run-initdb-dry

list:
	$(GO) list ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

quality: test vet

fmt:
	$(GO) fmt ./...

run-server:
	$(GO) run ./cmd/server server

run-server-dry:
	$(GO) run ./cmd/server server --dry-run

run-initdb:
	$(GO) run ./cmd/server initdb

run-initdb-dry:
	$(GO) run ./cmd/server initdb --dry-run
