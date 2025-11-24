build_server_out_folder = ./build/server_out
build_agent_out_folder = ./build/agent_out

bs_out = $(build_server_out_folder)/server
ba_out = $(build_agent_out_folder)/agent

dsn_db = postgres://postgres:admin54321@localhost:5678/postgres

src_server = ./cmd/server/main.go
src_agent = ./cmd/agent/main.go

build_server:
	rm -rf $(build_server_out_folder)
	mkdir -p $(build_server_out_folder)
	go build -o $(bs_out) $(src_server)

build_agent:
	rm -rf $(build_agent_out_folder)
	mkdir -p $(build_agent_out_folder)
	go build -o $(ba_out) $(src_agent)

build_all: build_agent build_server
	@echo "Builded agent and server"

run_server:
	go run cmd/server/main.go

run_agent:
	go run cmd/agent/main.go

test_by_bin: build_all
# 	rm -rf ./messages.log
# 	./metricstest_v2  -test.v -test.run=^TestIteration14 -source-path=. -agent-binary-path=$(ba_out) -binary-path=$(bs_out) -server-port=5555 -key=tmp -database-dsn=$(dsn_db) >> messages.log
	./metricstest_v2  -test.v -test.run=^TestIteration14 -source-path=. -agent-binary-path=$(ba_out) -binary-path=$(bs_out) -server-port=5555 -key=tmp -database-dsn=$(dsn_db) 

test_local:
	go test ./...

build_test_local_all: build_all test_local
	@echo "Full run build and test for server finished"

test_coverage:
	go test ./... -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out -o=./coverage.html