# LangGraph Go Showcases Makefile
.PHONY: help build build-all clean test fmt lint deps install run-* tidy check

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build output directory
BUILD_DIR=bin

# Showcases directories
SHOWCASES = BettaFish \
			gpt_researcher \
			health_insights_agent \
			pepolehub \
			deepagents \
			deerflow \
			langmanus \
			profile \
			ai-pdf-chatbot/backend

# Default target
help: ## Display this help message
	@echo "LangGraph Go Showcases - Available targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m\033[0m"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dependencies

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

##@ Building

build-all: ## Build all showcases
	@echo "Building all showcases..."
	@mkdir -p $(BUILD_DIR)
	@for dir in $(SHOWCASES); do \
		echo "Building $$dir..."; \
		if [ "$$dir" = "ai-pdf-chatbot/backend" ]; then \
			output_name="ai-pdf-chatbot"; \
		else \
			output_name=$$(basename $$dir); \
		fi; \
		(cd $$dir && $(GOBUILD) -o ../../$(BUILD_DIR)/$$output_name -v .) || exit 1; \
	done
	@echo "All builds completed. Binaries are in ./$(BUILD_DIR)/"

build-%: ## Build specific showcase (e.g., make build-BettaFish)
	@showcase_name=$*; \
	if [ "$$showcase_name" = "ai-pdf-chatbot" ]; then \
		dir="ai-pdf-chatbot/backend"; \
	else \
		dir=$$showcase_name; \
	fi; \
	if [ ! -d "$$dir" ]; then \
		echo "Error: Showcase '$$dir' not found"; \
		exit 1; \
	fi; \
	echo "Building $$showcase_name..."; \
	mkdir -p $(BUILD_DIR); \
	(cd $$dir && $(GOBUILD) -o ../../$(BUILD_DIR)/$$showcase_name -v .) || exit 1; \
	echo "Build completed: ./$(BUILD_DIR)/$$showcase_name"

install: ## Install all showcases to $GOPATH/bin
	@echo "Installing all showcases..."
	@for dir in $(SHOWCASES); do \
		echo "Installing $$dir..."; \
		(cd $$dir && $(GOCMD) install -v .) || exit 1; \
	done
	@echo "Installation completed."

##@ Running

run-BettaFish: ## Run BettaFish showcase
	@echo "Running BettaFish..."
	@cd BettaFish && $(GOCMD) run .

run-gpt_researcher: ## Run GPT Researcher showcase
	@echo "Running GPT Researcher..."
	@cd gpt_researcher && $(GOCMD) run .

run-health_insights_agent: ## Run Health Insights Agent showcase
	@echo "Running Health Insights Agent..."
	@cd health_insights_agent && $(GOCMD) run .

run-pepolehub: ## Run PeopleHub showcase
	@echo "Running PeopleHub..."
	@cd pepolehub && $(GOCMD) run .

run-deepagents: ## Run DeepAgents showcase
	@echo "Running DeepAgents..."
	@cd deepagents && $(GOCMD) run .

run-deerflow: ## Run DeerFlow showcase
	@echo "Running DeerFlow..."
	@cd deerflow && $(GOCMD) run .

run-langmanus: ## Run LangManus showcase
	@echo "Running LangManus..."
	@cd langmanus && $(GOCMD) run .

run-profile: ## Run Profile showcase
	@echo "Running Profile..."
	@cd profile && $(GOCMD) run .

run-ai-pdf-chatbot: ## Run AI PDF Chatbot showcase
	@echo "Running AI PDF Chatbot..."
	@cd ai-pdf-chatbot/backend && $(GOCMD) run .

##@ Testing

test: ## Run tests for all showcases
	@echo "Running tests..."
	@for dir in $(SHOWCASES); do \
		if [ -n "$$(find $$dir -name '*_test.go' -print -quit)" ]; then \
			echo "Testing $$dir..."; \
			(cd $$dir && $(GOTEST) -v ./...) || exit 1; \
		fi; \
	done
	@echo "All tests passed."

test-%: ## Run tests for specific showcase (e.g., make test-BettaFish)
	@showcase_name=$*; \
	if [ "$$showcase_name" = "ai-pdf-chatbot" ]; then \
		dir="ai-pdf-chatbot/backend"; \
	else \
		dir=$$showcase_name; \
	fi; \
	if [ ! -d "$$dir" ]; then \
		echo "Error: Showcase '$$dir' not found"; \
		exit 1; \
	fi; \
	echo "Testing $$showcase_name..."; \
	(cd $$dir && $(GOTEST) -v ./...) || exit 1

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@for dir in $(SHOWCASES); do \
		if [ -n "$$(find $$dir -name '*_test.go' -print -quit)" ]; then \
			echo "Testing $$dir with coverage..."; \
			(cd $$dir && $(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...) || exit 1; \
		fi; \
	done

##@ Code Quality

fmt: ## Format code using gofmt
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Code formatted."

vet: ## Run go vet
	@echo "Running go vet..."
	@for dir in $(SHOWCASES); do \
		echo "Vetting $$dir..."; \
		(cd $$dir && $(GOVET) ./...) || exit 1; \
	done
	@echo "Vet completed."

lint: ## Run golangci-lint if available
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it from: https://golangci-lint.run/welcome/install/"; \
		echo "Running go vet instead..."; \
		$(MAKE) vet; \
	fi

check: fmt vet ## Run fmt and vet

##@ Cleanup

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@find . -name "coverage.txt" -type f -delete
	@$(GOCLEAN)
	@echo "Clean completed."

clean-all: clean ## Clean all artifacts including dependencies cache
	@echo "Cleaning dependency cache..."
	@$(GOCLEAN) -modcache
	@echo "All cleaned."

##@ Information

list: ## List all showcases
	@echo "Available showcases:"
	@for dir in $(SHOWCASES); do \
		if [ "$$dir" = "ai-pdf-chatbot/backend" ]; then \
			echo "  - ai-pdf-chatbot"; \
		else \
			echo "  - $$dir"; \
		fi; \
	done

version: ## Show Go version
	@$(GOCMD) version

env: ## Show Go environment
	@$(GOCMD) env
