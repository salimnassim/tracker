.PHONY: sqlc-generate test build

sqlc-generate:
	docker run --rm -v "$$(pwd)":/src -w /src sqlc/sqlc generate

test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -o ./tracker ./cmd
