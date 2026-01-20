# Directories
BUILD_DIR := ./build
COVERAGE_DIR := ./go_test

# Binary output paths
BS_OUT := $(BUILD_DIR)/server_out/server
BA_OUT := $(BUILD_DIR)/agent_out/agent

# Source files
SRC_SERVER := ./cmd/server/main.go
SRC_AGENT := ./cmd/agent/main.go

# Database DSN
DSN_DB := postgres://postgres:admin54321localhost:5678/postgres

# Coverage files
COVERAGE_OUT := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html


# Recreate build directory
.PHONY: recreate_build_dir
recreate_build_dir:
	echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/server_out
	mkdir -p $(BUILD_DIR)/agent_out
	echo "Build directory prepared"

# Recreate coverage directory
.PHONY: recreate_coverage_dir
recreate_coverage_dir:
	echo "Cleaning coverage directory..."
	rm -rf $(COVERAGE_DIR)
	mkdir -p $(COVERAGE_DIR)
	echo "Coverage directory prepared"


# Build server
build_server: recreate_build_dir
	echo "Building server..."
	go build -o $(BS_OUT) $(SRC_SERVER)
	echo "Server built: $(BS_OUT)"

# Build agent
build_agent: recreate_build_dir
	echo "Building agent..."
	go build -o $(BA_OUT) $(SRC_AGENT)
	echo "Agent built: $(BA_OUT)"

# Build both
build_all: build_agent build_server
	echo "Build complete: server and agent"

# Run server
run_server:
	go run $(SRC_SERVER)

# Run agent
run_agent:
	go run $(SRC_AGENT)

# Run binary tests
test_by_bin: build_all
	./metricstest_v2 \
		-test.v \
		-test.run=^TestIteration16 \
		-source-path=. \
		-agent-binary-path=$(BA_OUT) \
		-binary-path=$(BS_OUT) \
		-server-port=5555 \
		-key=tmp \
		-database-dsn=$(DSN_DB)

# Run local tests
test_local:
	go test ./...

# Full build and test
build_test_local_all: build_all test_local
	echo "Build and local tests completed"

# Test coverage
test_coverage: recreate_coverage_dir
	echo "Running tests with coverage..."
	go test ./... -coverprofile=$(COVERAGE_OUT) -covermode=atomic
	echo "Generating coverage report..."
	go tool cover -func=$(COVERAGE_OUT)
	go tool cover -html=$(COVERAGE_OUT) -o=$(COVERAGE_HTML)
	echo "Coverage report generated: $(COVERAGE_HTML)"

# Clean all artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	echo "Clean completed: removed $(BUILD_DIR) and $(COVERAGE_DIR)"
