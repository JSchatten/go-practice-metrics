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


# Build tags
BUILD_VERSION="1.2.3"
# BUILD_DATE=today_hehe
# Попробуем вытащить из шелла
BUILD_DATE    := $(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_COMMIT  := $(shell git rev-parse HEAD)

default: clean recreate_dirs gen_reset test_coverage go_doc_full go_staticlint build_all
	echo "Full build"

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

recreate_dirs: recreate_build_dir recreate_coverage_dir 
	echo "recreate directories"

# Generate files
build_reset:
	go build -o $(BUILD_DIR)/resetgen ./cmd/reset/main.go

gen_reset: build_reset
	echo "Generate reset files..."
	$(BUILD_DIR)/resetgen

# Build server
build_server:
	echo "Building server..."
	go build \
		-ldflags "\
			-X 'main.buildVersion=$(BUILD_VERSION)' \
			-X 'main.buildDate=\"$(BUILD_DATE)\"' \
			-X 'main.buildCommit=\"$(BUILD_COMMIT)\"' \
		" \
		-o $(BS_OUT) $(SRC_SERVER)
	echo "Server built: $(BS_OUT)"

# Build agent
build_agent:
	echo "Building agent..."
	go build \
		-ldflags "\
			-X 'main.buildVersion=$(BUILD_VERSION)' \
			-X 'main.buildDate=\"$(BUILD_DATE)\"' \
			-X 'main.buildCommit=\"$(BUILD_MESSAGE)\"' \
		" \
		-o $(BA_OUT) $(SRC_AGENT)
	echo "Agent built: $(BA_OUT)"

build_linter:
	go build -o $(BUILD_DIR)/staticlint cmd/staticlint/main.go

# Build both
build_all: build_agent build_server
	echo "Build complete: server and agent"

# Run server
run_server:
	go run $(SRC_SERVER)

run_server_x:
	$(BS_OUT)

# Run agent
run_agent:
	go run $(SRC_AGENT)

run_agent_x:
	go run $(BA_OUT)

# Run binary tests
test_by_bin: build_all
	./metricstest_v2 \
		-test.v \
		-test.run=^TestIteration19 \
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
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	echo "Clean completed: removed $(BUILD_DIR) and $(COVERAGE_DIR)"

go_md_doc:
	mkdir -p docs
#	Need gon install github.com/robertkrimen/godocdown/godocdown@latest
	godocdown ./internal/model > docs/model.md
	godocdown ./internal/handler > docs/handler.md

go_doc:
	go test -v ./internal/model -run Example
	go doc -all model.Metrics
	go doc -all handler.ValueHandler

go_doc_full: go_md_doc go_doc
# 	echo "===\nDocs shown and generated"

# 	./staticlint ./cmd/... ./internal/... ./pkg/...
# 	go build -o staticlint cmd/staticlint/main.go
go_staticlint: build_linter
	go vet -vettool=$(BUILD_DIR)/staticlint ./cmd/... ./internal/... ./pkg/...

gen_crypto_keys:
	openssl genrsa -out private_test.pem 5096
	openssl rsa -in private_test.pem -pubout -out public_test.pem