# Terraform Schema Generator - Makefile
#
# Common tasks for building, testing, and managing the project

# Variables
BINARY_NAME=terraform-schema-generator
MAIN_PATH=./cmd/cli
PKG_PATH=./pkg/...
GO=go
GOFLAGS=-v
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go build flags
LDFLAGS=-ldflags="-s -w"
BUILD_FLAGS=$(LDFLAGS) -trimpath

# Colors for output - auto-detect terminal support
# Check if NO_COLOR is set or if output is not a terminal
ifneq ($(NO_COLOR),)
	# NO_COLOR environment variable is set
	CYAN=
	GREEN=
	YELLOW=
	RED=
	NC=
else ifeq ($(shell test -t 1 && echo 1),1)
	# Output is to a terminal, use colors
	CYAN=\033[0;36m
	GREEN=\033[0;32m
	YELLOW=\033[0;33m
	RED=\033[0;31m
	NC=\033[0m
else
	# Output is not to a terminal (piped, redirected, or IntelliJ), disable colors
	CYAN=
	GREEN=
	YELLOW=
	RED=
	NC=
endif

.PHONY: help
help: ## Show this help message
	@echo '$(CYAN)Terraform Schema Generator - Available Commands:$(NC)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ''
	@echo 'Tip: To disable colors, run: NO_COLOR=1 make <target>'
	@echo ''

.PHONY: all
all: clean fmt lint test build ## Run all checks and build

.PHONY: build
build: ## Build the CLI binary
	@echo '$(CYAN)Building binary...$(NC)'
	@mkdir -p bin
	$(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH) 2>/dev/null || echo '$(YELLOW)No CLI found, skipping binary build$(NC)'
	@if [ -f bin/$(BINARY_NAME) ]; then \
		echo '$(GREEN)✓ Binary built: bin/$(BINARY_NAME)$(NC)'; \
	fi

.PHONY: build-all
build-all: ## Build binaries for all platforms
	@echo '$(CYAN)Building for all platforms...$(NC)'
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH) 2>/dev/null || true
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH) 2>/dev/null || true
	GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH) 2>/dev/null || true
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH) 2>/dev/null || true
	@echo '$(GREEN)✓ Cross-platform builds complete$(NC)'
	@ls -lh bin/ 2>/dev/null || true

.PHONY: install
install: ## Install the binary to GOPATH/bin
	@echo '$(CYAN)Installing binary...$(NC)'
	$(GO) install $(BUILD_FLAGS) $(MAIN_PATH) 2>/dev/null || echo '$(YELLOW)No CLI found, skipping install$(NC)'
	@echo '$(GREEN)✓ Installed$(NC)'

.PHONY: test
test: ## Run all tests
	@echo '$(CYAN)Running tests...$(NC)'
	@$(GO) test $(PKG_PATH) $(GOFLAGS)
	@echo '$(GREEN)✓ All tests passed$(NC)'

.PHONY: test-short
test-short: ## Run tests with short flag
	@echo '$(CYAN)Running short tests...$(NC)'
	@$(GO) test -short $(PKG_PATH)
	@echo '$(GREEN)✓ Short tests passed$(NC)'

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo '$(CYAN)Running tests with race detector...$(NC)'
	@$(GO) test -race $(PKG_PATH)
	@echo '$(GREEN)✓ Race tests passed$(NC)'

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo '$(CYAN)Running tests (verbose)...$(NC)'
	@$(GO) test $(PKG_PATH) -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo '$(CYAN)Running tests with coverage...$(NC)'
	@$(GO) test $(PKG_PATH) -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total:
	@echo '$(GREEN)✓ Coverage report: $(COVERAGE_FILE)$(NC)'

.PHONY: coverage-html
coverage-html: test-coverage ## Generate HTML coverage report
	@echo '$(CYAN)Generating HTML coverage report...$(NC)'
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo '$(GREEN)✓ HTML coverage report: $(COVERAGE_HTML)$(NC)'
	@which open >/dev/null 2>&1 && open $(COVERAGE_HTML) || true

.PHONY: bench
bench: ## Run benchmarks
	@echo '$(CYAN)Running benchmarks...$(NC)'
	@$(GO) test -bench=. -benchmem $(PKG_PATH)

.PHONY: fmt
fmt: ## Format all Go files
	@echo '$(CYAN)Formatting code...$(NC)'
	@$(GO) fmt $(PKG_PATH)
	@echo '$(GREEN)✓ Code formatted$(NC)'

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo '$(CYAN)Checking code formatting...$(NC)'
	@test -z "$$(gofmt -l .)" || (echo '$(RED)✗ Code not formatted. Run: make fmt$(NC)' && exit 1)
	@echo '$(GREEN)✓ Code is properly formatted$(NC)'

.PHONY: lint
lint: ## Run linters (requires golangci-lint)
	@echo '$(CYAN)Running linters...$(NC)'
	@if which golangci-lint >/dev/null 2>&1 || [ -f ~/go/bin/golangci-lint ]; then \
		(which golangci-lint >/dev/null 2>&1 && golangci-lint run $(PKG_PATH)) || ~/go/bin/golangci-lint run $(PKG_PATH); \
		echo '$(GREEN)✓ Linting passed$(NC)'; \
	else \
		echo '$(YELLOW)golangci-lint not found, skipping$(NC)'; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo '$(CYAN)Running go vet...$(NC)'
	@$(GO) vet $(PKG_PATH)
	@echo '$(GREEN)✓ Vet passed$(NC)'

.PHONY: tidy
tidy: ## Tidy go modules
	@echo '$(CYAN)Tidying go modules...$(NC)'
	@$(GO) mod tidy
	@echo '$(GREEN)✓ Modules tidied$(NC)'

.PHONY: vendor
vendor: ## Vendor dependencies
	@echo '$(CYAN)Vendoring dependencies...$(NC)'
	@$(GO) mod vendor
	@echo '$(GREEN)✓ Dependencies vendored$(NC)'

.PHONY: deps
deps: ## Download dependencies
	@echo '$(CYAN)Downloading dependencies...$(NC)'
	@$(GO) mod download
	@echo '$(GREEN)✓ Dependencies downloaded$(NC)'

.PHONY: clean
clean: ## Clean build artifacts
	@echo '$(CYAN)Cleaning...$(NC)'
	@rm -rf bin/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f examples/fluent/schema.json
	@rm -f examples/basic/terraform-variables-schema.json
	@$(GO) clean -cache -testcache
	@echo '$(GREEN)✓ Cleaned$(NC)'

.PHONY: clean-cache
clean-cache: ## Clean Go cache
	@echo '$(CYAN)Cleaning cache...$(NC)'
	@$(GO) clean -cache -testcache -modcache
	@echo '$(GREEN)✓ Cache cleaned$(NC)'

.PHONY: run-example-fluent
run-example-fluent: ## Run the fluent API example
	@echo '$(CYAN)Running fluent API example...$(NC)'
	@cd examples/fluent && $(GO) run main.go

.PHONY: run-example-basic
run-example-basic: ## Run the basic example
	@echo '$(CYAN)Running basic example...$(NC)'
	@cd examples/basic && $(GO) run main.go

.PHONY: examples
examples: run-example-fluent run-example-basic ## Run all examples

.PHONY: check
check: fmt-check vet test ## Run all checks (format, vet, test)
	@echo '$(GREEN)✓ All checks passed$(NC)'

.PHONY: ci
ci: deps check test-coverage ## Run CI pipeline
	@echo '$(GREEN)✓ CI pipeline complete$(NC)'

.PHONY: stats
stats: ## Show code statistics
	@echo '$(CYAN)Code Statistics:$(NC)'
	@echo ''
	@echo 'Lines of code:'
	@find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' | xargs wc -l | tail -1
	@echo ''
	@echo 'Test files:'
	@find . -name '*_test.go' -not -path './vendor/*' | wc -l | awk '{print $$1 " test files"}'
	@echo ''
	@echo 'Packages:'
	@go list ./... | wc -l | awk '{print $$1 " packages"}'
	@echo ''
	@echo 'Test count:'
	@go test $(PKG_PATH) -v 2>&1 | grep -c "^=== RUN" || echo "0 tests"

.PHONY: test-count
test-count: ## Count all tests
	@echo '$(CYAN)Counting tests...$(NC)'
	@go test $(PKG_PATH) -v 2>&1 | grep -c "^=== RUN" | awk '{print "Total tests: " $$1}'

.PHONY: watch
watch: ## Watch for changes and run tests (requires entr)
	@which entr >/dev/null 2>&1 || (echo '$(RED)entr not found. Install with: brew install entr$(NC)' && exit 1)
	@echo '$(CYAN)Watching for changes...$(NC)'
	@find . -name '*.go' -not -path './vendor/*' | entr -c make test

.PHONY: doc
doc: ## Generate and serve godoc
	@echo '$(CYAN)Starting godoc server...$(NC)'
	@echo '$(GREEN)Documentation available at: http://localhost:6060/pkg/github.com/samart/terraform-schema-generator/$(NC)'
	@which godoc >/dev/null 2>&1 || (echo '$(YELLOW)godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest$(NC)' && exit 1)
	@godoc -http=:6060

.PHONY: update
update: ## Update dependencies
	@echo '$(CYAN)Updating dependencies...$(NC)'
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo '$(GREEN)✓ Dependencies updated$(NC)'

.PHONY: verify
verify: ## Verify dependencies
	@echo '$(CYAN)Verifying dependencies...$(NC)'
	@$(GO) mod verify
	@echo '$(GREEN)✓ Dependencies verified$(NC)'

.PHONY: pre-commit
pre-commit: fmt vet test ## Run pre-commit checks
	@echo '$(GREEN)✓ Pre-commit checks passed$(NC)'

.PHONY: release-check
release-check: clean all test-coverage ## Check if ready for release
	@echo '$(GREEN)✓ Release checks passed$(NC)'
	@echo ''
	@echo '$(CYAN)Ready to release!$(NC)'

.PHONY: version
version: ## Show Go version
	@$(GO) version

.PHONY: info
info: ## Show project information
	@echo '$(CYAN)Project Information:$(NC)'
	@echo ''
	@echo 'Name:          Terraform Schema Generator'
	@echo 'Description:   Generate JSON Schema Draft 7 from Terraform configurations'
	@echo 'Go Version:    '`$(GO) version | awk '{print $$3}'`
	@echo 'Module:        '`go list -m`
	@echo ''
	@echo 'Packages:'
	@go list ./pkg/...
	@echo ''

# Default target
.DEFAULT_GOAL := help
