.PHONY: sqlc-generate test build ci

sqlc-generate:
	docker run --rm -v "$$(pwd)":/src -w /src sqlc/sqlc generate

test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -o ./tracker ./cmd

ci:
	unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	go build ./...
	go vet ./...
	go test -race ./...
	go mod tidy
	git diff --exit-code -- go.mod go.sum
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
