# Text2SVG // CAM & CAD Vector Studio
# Makefile

BINARY_NAME ?= text2svg
MAIN_PKG    ?= .
GO          ?= go
PORT        ?= 8080
DIST_DIR    ?= dist

# Build flags
CGO_ENABLED ?= 0
LDFLAGS     ?= -s -w

.PHONY: all build run dev test test-race coverage fmt vet lint tidy install kill build-all clean help

# Default target
all: build

## Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  \033[36m%-15s\033[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Compile the binary for the current system
build:
	@printf '\033[1;36m==>\033[0m Building %s...\n' "$(BINARY_NAME)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PKG)
	@printf '\033[1;32m✓\033[0m Built: ./%s\n' "$(BINARY_NAME)"

## Build and run the interactive Web Studio
run: build
	@printf '\033[1;36m==>\033[0m Launching %s on port %s...\n' "$(BINARY_NAME)" "$(PORT)"
	./$(BINARY_NAME) --port $(PORT)

## Run the Web Studio without auto-opening browser (useful for development)
dev:
	@printf '\033[1;36m==>\033[0m Starting dev server on port %s (no browser auto-launch)...\n' "$(PORT)"
	$(GO) run $(MAIN_PKG) --port $(PORT) --no-browser

## Run unit tests
test:
	@printf '\033[1;36m==>\033[0m Running tests...\n'
	$(GO) test -v ./...

## Run unit tests with race detection
test-race:
	@printf '\033[1;36m==>\033[0m Running tests with race detector...\n'
	$(GO) test -v -race ./...

## Run tests and generate HTML code coverage report
coverage:
	@printf '\033[1;36m==>\033[0m Generating test coverage...\n'
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@printf '\033[1;32m✓\033[0m Coverage report written to coverage.html\n'

## Format Go source files
fmt:
	@printf '\033[1;36m==>\033[0m Formatting source code...\n'
	$(GO) fmt ./...

## Run go vet static analysis
vet:
	@printf '\033[1;36m==>\033[0m Running go vet...\n'
	$(GO) vet ./...

## Run linter (golangci-lint if installed, fallback to go vet)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		printf '\033[1;36m==>\033[0m Running golangci-lint...\n'; \
		golangci-lint run; \
	else \
		printf '\033[1;33m!\033[0m golangci-lint not installed; running go vet fallback...\n'; \
		$(GO) vet ./...; \
	fi

## Prune and resolve go.mod dependencies
tidy:
	@printf '\033[1;36m==>\033[0m Tidying module dependencies...\n'
	$(GO) mod tidy

## Install binary into GOPATH/bin
install:
	@printf '\033[1;36m==>\033[0m Installing to GOPATH/bin...\n'
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install -ldflags="$(LDFLAGS)" $(MAIN_PKG)
	@printf '\033[1;32m✓\033[0m Installed successfully.\n'

## Cross-compile release binaries for multiple OS and architectures
build-all:
	@printf '\033[1;36m==>\033[0m Cross-compiling release binaries into %s/...\n' "$(DIST_DIR)"
	@mkdir -p $(DIST_DIR)
	# macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	# macOS Intel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	# Linux 64-bit
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	# Linux ARM64 (e.g. Raspberry Pi)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	# Windows 64-bit
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)
	@printf '\033[1;32m✓\033[0m Cross-compilation complete:\n'
	@ls -lh $(DIST_DIR)/

## Kill running text2svg processes and free the port
kill:
	@PIDS=$$(pgrep -f '/$(BINARY_NAME)$$|^\./$(BINARY_NAME)' 2>/dev/null || true); \
	PORT_PIDS=$$(lsof -tiTCP:$(PORT) -sTCP:LISTEN 2>/dev/null || true); \
	ALL=$$(echo "$$PIDS $$PORT_PIDS" | tr ' ' '\n' | sort -u | grep -v '^$$' || true); \
	if [ -n "$$ALL" ]; then \
	  printf '\033[1;33m●\033[0m Stopping process(es): %s\n' "$$ALL"; \
	  kill $$ALL 2>/dev/null || true; sleep 1; \
	  for p in $$ALL; do kill -0 $$p 2>/dev/null && kill -9 $$p 2>/dev/null || true; done; \
	  printf '\033[1;32m✓\033[0m Stopped.\n'; \
	else \
	  printf '\033[2m·\033[0m No %s process found on port %s.\n' "$(BINARY_NAME)" "$(PORT)"; \
	fi

## Remove compiled binaries, coverage files, and dist directory
clean:
	@printf '\033[1;36m==>\033[0m Cleaning build artifacts...\n'
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe coverage.out coverage.html
	@rm -rf $(DIST_DIR)
	@printf '\033[1;32m✓\033[0m Clean complete.\n'
