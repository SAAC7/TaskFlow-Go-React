.PHONY: run test vet build tidy

run:   ## Start the API (reads .env)
	go run ./cmd/api

test:  ## Run unit tests
	go test ./... -race -cover

vet:   ## Static analysis
	go vet ./...

build: ## Build the binary into ./bin
	go build -o bin/api ./cmd/api

tidy:  ## Download dependencies and update go.mod/go.sum
	go mod tidy
