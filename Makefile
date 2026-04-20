GO ?= go

.PHONY: list test vet quality fmt run run-dry db-migrate db-migrate-dry

list:
	$(GO) list ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

quality: test vet

fmt:
	$(GO) fmt ./...

run:
	$(GO) run ./cmd run

run-dry:
	$(GO) run ./cmd run --dry-run

db-migrate:
	$(GO) run ./cmd db migrate

db-migrate-dry:
	$(GO) run ./cmd db migrate --dry-run
