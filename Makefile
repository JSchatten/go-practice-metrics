build_server_out = ./build/server_out


build_server:
	rm -rf $(build_server_out)
	mkdir -p $(build_server_out)
	go build -o $(build_server_out)/server ./cmd/server/main.go

run_server:
	go run cmd/server/main.go

test_server:
	./metricstest  -test.v -test.run=^TestIteration1$ -server-binary-path=./build/server_out/server

server_build_test: build_server test_server
	@echo "Full run fo server finished"
