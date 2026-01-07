# LangGraph Go Showcases Makefile
.PHONY: all help build build-all clean test fmt fmt-check lint deps install run-* tidy check modernize

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Colors
COLOR_RESET   = \033[0m
COLOR_BOLD    = \033[1m
COLOR_RED     = \033[31m
COLOR_GREEN   = \033[32m
COLOR_YELLOW  = \033[33m
COLOR_BLUE    = \033[34m
COLOR_CYAN    = \033[36m

# Build output directory
BUILD_DIR=bin

# Showcases directories
SHOWCASES = BettaFish \
			gpt_researcher \
			health_insights_agent \
			pepolehub \
			deepagents \
			Insight \
			langmanus \
			profile \
			ai-pdf-chatbot/backend

# Default target: run all checks and build
all: deps fmt-check vet lint test build-all ## Run all checks and build (default target)
	@echo "$(COLOR_GREEN)✓ All checks passed and build completed!$(COLOR_RESET)"

help: ## Display this help message
	@echo "$(COLOR_BOLD)LangGraph Go Showcases - Available targets:$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m\033[0m"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dependencies

deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@$(GOMOD) download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

tidy: ## Tidy go.mod and go.sum
	@echo "$(COLOR_BLUE)Tidying dependencies...$(COLOR_RESET)"
	@$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Dependencies tidied$(COLOR_RESET)"

##@ Building

build-all: ## Build all showcases
	@echo "$(COLOR_BLUE)Building all showcases...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@for dir in $(SHOWCASES); do \
		if [ "$$dir" = "ai-pdf-chatbot/backend" ]; then \
			output_name="ai-pdf-chatbot"; \
		else \
			output_name=$$(basename $$dir); \
		fi; \
		echo "$(COLOR_CYAN)  Building $$output_name...$(COLOR_RESET)"; \
		(cd $$dir && $(GOBUILD) -o ../../$(BUILD_DIR)/$$output_name -v .) || exit 1; \
	done
	@echo "$(COLOR_GREEN)✓ All builds completed. Binaries are in ./$(BUILD_DIR)/$(COLOR_RESET)"

build-%: ## Build specific showcase (e.g., make build-BettaFish)
	@showcase_name=$*; \
	if [ "$$showcase_name" = "ai-pdf-chatbot" ]; then \
		dir="ai-pdf-chatbot/backend"; \
	else \
		dir=$$showcase_name; \
	fi; \
	if [ ! -d "$$dir" ]; then \
		echo "$(COLOR_RED)Error: Showcase '$$dir' not found$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	echo "$(COLOR_BLUE)Building $$showcase_name...$(COLOR_RESET)"; \
	mkdir -p $(BUILD_DIR); \
	(cd $$dir && $(GOBUILD) -o ../../$(BUILD_DIR)/$$showcase_name -v .) || exit 1; \
	echo "$(COLOR_GREEN)✓ Build completed: ./$(BUILD_DIR)/$$showcase_name$(COLOR_RESET)"

install: ## Install all showcases to $GOPATH/bin
	@echo "$(COLOR_BLUE)Installing all showcases...$(COLOR_RESET)"
	@for dir in $(SHOWCASES); do \
		echo "$(COLOR_CYAN)  Installing $$dir...$(COLOR_RESET)"; \
		(cd $$dir && $(GOCMD) install -v .) || exit 1; \
	done
	@echo "$(COLOR_GREEN)✓ Installation completed$(COLOR_RESET)"

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

run-Insight: ## Run Insight showcase
	@echo "Running Insight..."
	@cd Insight && $(GOCMD) run .

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
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@for dir in $(SHOWCASES); do \
		if [ -n "$$(find $$dir -name '*_test.go' -print -quit)" ]; then \
			echo "$(COLOR_CYAN)  Testing $$dir...$(COLOR_RESET)"; \
			(cd $$dir && $(GOTEST) -v ./...) || exit 1; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✓ All tests passed$(COLOR_RESET)"

test-%: ## Run tests for specific showcase (e.g., make test-BettaFish)
	@showcase_name=$*; \
	if [ "$$showcase_name" = "ai-pdf-chatbot" ]; then \
		dir="ai-pdf-chatbot/backend"; \
	else \
		dir=$$showcase_name; \
	fi; \
	if [ ! -d "$$dir" ]; then \
		echo "$(COLOR_RED)Error: Showcase '$$dir' not found$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	echo "$(COLOR_BLUE)Testing $$showcase_name...$(COLOR_RESET)"; \
	(cd $$dir && $(GOTEST) -v ./...) || exit 1; \
	echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	@for dir in $(SHOWCASES); do \
		if [ -n "$$(find $$dir -name '*_test.go' -print -quit)" ]; then \
			echo "$(COLOR_CYAN)  Testing $$dir with coverage...$(COLOR_RESET)"; \
			(cd $$dir && $(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...) || exit 1; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✓ Coverage reports generated$(COLOR_RESET)"

##@ Code Quality

fmt: ## Format code using gofmt
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GOFMT) -s -w .
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

fmt-check: ## Check if code is formatted (without modifying)
	@echo "$(COLOR_BLUE)Checking code formatting...$(COLOR_RESET)"
	@FMT_OUTPUT=$$($(GOFMT) -l .); \
	if [ -n "$$FMT_OUTPUT" ]; then \
		echo "$(COLOR_RED)✗ The following files need formatting:$(COLOR_RESET)"; \
		echo "$$FMT_OUTPUT" | sed 's/^/  /'; \
		echo "$(COLOR_YELLOW)Run 'make fmt' to fix formatting$(COLOR_RESET)"; \
		exit 1; \
	else \
		echo "$(COLOR_GREEN)✓ All code is properly formatted$(COLOR_RESET)"; \
	fi

vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@for dir in $(SHOWCASES); do \
		echo "$(COLOR_CYAN)  Vetting $$dir...$(COLOR_RESET)"; \
		(cd $$dir && $(GOVET) ./...) || exit 1; \
	done
	@echo "$(COLOR_GREEN)✓ Vet completed$(COLOR_RESET)"

lint: ## Run golangci-lint if available
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(COLOR_BLUE)Running golangci-lint...$(COLOR_RESET)"; \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint completed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not installed. Install it from: https://golangci-lint.run/welcome/install/$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Running go vet instead...$(COLOR_RESET)"; \
		$(MAKE) vet; \
	fi

modernize: ## Run modernize to apply fixes to all packages
	@echo "$(COLOR_BLUE)Running modernize...$(COLOR_RESET)"
	@if command -v modernize >/dev/null 2>&1; then \
		modernize -fix -test ./...; \
		echo "$(COLOR_GREEN)✓ Modernize completed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)modernize not installed. Install it with: go install golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest$(COLOR_RESET)"; \
		exit 1; \
	fi

check: fmt vet ## Run fmt and vet
	@echo "$(COLOR_GREEN)✓ All checks completed$(COLOR_RESET)"

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@find . -name "coverage.txt" -type f -delete
	@$(GOCLEAN)
	@echo "$(COLOR_GREEN)✓ Clean completed$(COLOR_RESET)"

clean-all: clean ## Clean all artifacts including dependencies cache
	@echo "$(COLOR_BLUE)Cleaning dependency cache...$(COLOR_RESET)"
	@$(GOCLEAN) -modcache
	@echo "$(COLOR_GREEN)✓ All cleaned$(COLOR_RESET)"

##@ Information

list: ## List all showcases
	@echo "$(COLOR_BOLD)Available showcases:$(COLOR_RESET)"
	@for dir in $(SHOWCASES); do \
		if [ "$$dir" = "ai-pdf-chatbot/backend" ]; then \
			echo "  $(COLOR_CYAN)•$(COLOR_RESET) ai-pdf-chatbot"; \
		else \
			echo "  $(COLOR_CYAN)•$(COLOR_RESET) $$dir"; \
		fi; \
	done

version: ## Show Go version
	@$(GOCMD) version

env: ## Show Go environment
	@$(GOCMD) env
