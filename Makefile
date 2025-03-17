# Define variables
BINARY_NAME=dish-dispatcher
BUILD_DIR=bin
SRC_DIR=cmd/server
MAIN_FILE=$(SRC_DIR)/main.go

# Default target
.PHONY: all
all: test build run

# Run tests
.PHONY: test
test:
	go test ./internal/... -cover -race -v

# Build the binary
.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Run the application
.PHONY: run
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)