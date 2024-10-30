.DEFAULT_GOAL := help
BINARY_NAME :=  vojvoj

build: ## Build crossplatform go application
		GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin main.go
		GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
		GOARCH=amd64 GOOS=window go build -o ${BINARY_NAME}-windows main.go

test: ## Run Tests
		go test ./... -coverprofile=coverage.out
.PHONY: test

dep: ## Install and Update Dependencies
		go mod download
		go mod tidy
		go mod verify

lint: ## lint files. currently linting .go and .sql
	golangci-lint run --out-format=colored-line-number -c .golangci.yaml --fix
	sqlfluff fix --config ./scripts/generate/sqlfluff.ini --force
	sqlfluff lint --config ./scripts/generate/sqlfluff.ini
.PHONY: lint

rewrite: scripts/comby-rules ## apply comby rewrite rules to go files
	comby -timeout 10 -review -matcher '.go' -templates scripts/comby-rules -jobs 5 -extensions '.go'
.PHONY: rewrite


help: ## help
		@echo "Available commands"
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run: ## Run go application for dev environment in local
	go run main.go -conf configuration

deploy-dev: ## Deploy application in dev environment
	./scripts/deployment/deploy.sh ./scripts/deployment/dev


