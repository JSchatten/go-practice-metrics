build_server_out = ./build/server_out
build_agent_out = ./build/agent_out


build_server:
	rm -rf $(build_server_out)
	mkdir -p $(build_server_out)
	go build -o $(build_server_out)/server ./cmd/server/main.go

build_agent:
	rm -rf $(build_agent_out)
	mkdir -p $(build_agent_out)
	go build -o $(build_agent_out)/agent ./cmd/agent/main.go

run_server:
	go run cmd/server/main.go

run_agent:
	go run cmd/agent/main.go

test_server:
	./metricstest  -test.v -test.run=^TestIteration6$ -server-binary-path=./build/server_out/server

test_local:
	go test ./...

test_agent:
	./metricstest  -test.v -test.run=^TestIteration5$ -server-binary-path=./build/agent_out/agent

build_test_local_all: build_server build_agent test_local
	@echo "Full run build and test for server finished"

test_coverage:
	go test ./... -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out -o=./coverage.html