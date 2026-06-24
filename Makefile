PROJECT_NAME := "home-ip-monitor"
PKG := "github.com/a-castellano/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all build clean test tests tests_unit test_integration \
	test_domain test_domain_unit test_app test_app_unit test_config test_config_unit \
	test_ipinfodata test_ipinfodata_unit test_nslookup test_nslookup_unit \
	test_storage test_storage_unit test_notify test_notify_unit \
	coverage coverhtml lint race help

all: build

lint: ## Lint the files
	@go vet ./...

# Canonical targets (used by CI). Tag scheme:
#   unit_tests        -> only *_test.go (excludes *_integration_test.go)
#   integration_tests -> every test file (functional + unit), it is the superset
test: ## Run unit tests only
	@go test --tags=unit_tests -short ./...

test_integration: ## Run all tests: functional + unit
	@go test --tags=integration_tests -short ./...

# Friendly aliases
tests: test_integration ## Run all tests (functional + unit)

tests_unit: test ## Run unit tests only

# Per-module tests. Each <mod>_tests runs functional + unit for that module,
# each <mod>_unit_tests runs only its unit tests.
test_domain: ## Run domain tests
	@go test --tags=domain_tests -short ./...
test_domain_unit: ## Run domain unit tests only
	@go test --tags=domain_unit_tests -short ./...

test_app: ## Run app (use case) tests
	@go test --tags=app_tests -short ./...
test_app_unit: ## Run app unit tests only
	@go test --tags=app_unit_tests -short ./...

test_config: ## Run config tests
	@go test --tags=config_tests -short ./...
test_config_unit: ## Run config unit tests only
	@go test --tags=config_unit_tests -short ./...

test_ipinfodata: ## Run ipinfodata tests
	@go test --tags=ipinfodata_tests -short ./...
test_ipinfodata_unit: ## Run ipinfodata unit tests only
	@go test --tags=ipinfodata_unit_tests -short ./...

test_nslookup: ## Run nslookup tests
	@go test --tags=nslookup_tests -short ./...
test_nslookup_unit: ## Run nslookup unit tests only
	@go test --tags=nslookup_unit_tests -short ./...

test_storage: ## Run storage tests
	@go test --tags=storage_tests -short ./...
test_storage_unit: ## Run storage unit tests only
	@go test --tags=storage_unit_tests -short ./...

test_notify: ## Run notify tests
	@go test --tags=notify_tests -short ./...
test_notify_unit: ## Run notify unit tests only
	@go test --tags=notify_unit_tests -short ./...

race: ## Run data race detector
	@go test -race -short ./...

coverage: ## Generate global code coverage report
	./development/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	go tool cover -html=cover/coverage.report -o coverage.html;

build: ## Build the binary file
	@go build -v $(PKG)/cmd/$(PROJECT_NAME)

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
