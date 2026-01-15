.DEFAULT_GOAL = help
.PHONY: build dev-server dev-client lint lint-fix lint-fast install-deps clean 

PROJECT_DIR = $(shell pwd)
PROJECT_BIN = $(PROJECT_DIR)/bin
GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint

##
## Linters:
lint: ## - Run lint for checking code style
	$(GOLANGCI_LINT) run ./... --config=$(PROJECT_DIR)/.golangci.yaml 

lint-fix: ## - Run lint for auto-fix code style
	$(GOLANGCI_LINT) run ./... --fix --config=$(PROJECT_DIR)/.golangci.yaml

lint-fast: ## - Run lint with fast flag for partial checking code style
	$(GOLANGCI_LINT) run ./... --fast --config=$(PROJECT_DIR)/.golangci.yaml

##
## Run:
build: ## - Build server and client
build: build-server build-client

build-server: ## - Build server
	go build -o $(PROJECT_DIR)/cmd/server/main.go $(PROJECT_BIN)/debug/server 

build-client: ## - Build client
	go build -o $(PROJECT_DIR)/cmd/client/main.go $(PROJECT_BIN)/debug/client 

build-release: ## - Build server and client with release flags
build-release: build-release-server build-release-client

build-release-server: ## - Build server with release flags
	go CGO_ENABLED=0 GOOS=linux build -o $(PROJECT_DIR)/cmd/server/main.go $(PROJECT_BIN)/release/server-release

build-release-client: ## - Build client with release flags
	go CGO_ENABLED=0 GOOS=linux build -o $(PROJECT_DIR)/cmd/client/main.go $(PROJECT_BIN)/release/client-release

dev-server: ## - Dev run server
	go run cmd/server/main.go

dev-client: ## - Dev run client
	go run cmd/client/main.go

##
## Tests:
test: ## - Run tests
	go test -v ./...

##
## Other:
help: ## - Show help message
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/\\$$//' | awk 'BEGIN {FS = ":*[[:space:]]*##[[:space:]]*"}; \
	{ \
		if($$2 == "") \
			printf ""; \
		else if($$0 ~ /^#/) \
			printf "\n%s\n", $$2; \
		else if($$1 == "") \
			printf "     %-20s%s\n", "", $$2; \
		else \
			printf "\n    \033[34m%-20s\033[0m %s\n", $$1, $$2; \
	}'
	@echo -e "" # blank line at the end

clean: ## - Clear bin dir and git hooks
	rm -rf $(PROJECT_BIN) && rm -f .git/hooks/pre-commit

install-deps: ## - Install all deps in repo
install-deps: .install-linter .install-hooks

.install-linter: ## - (optional) Install golangci-lint in bin dir
	$(PROJECT_DIR)/scripts/install-golangci-lint.sh

.install-hooks: ## - (optional) Install pre commit hook
	$(PROJECT_DIR)/scripts/install-pre-commit-hook.sh

##
## Usage: make <command>
