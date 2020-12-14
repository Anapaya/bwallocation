.PHONY: mocks proto test build

build:
	go build -v -o bin/bwreserver ./cmd/bwreserver

mocks:
	mockgen github.com/anapaya/bwallocation/proto/bw_allocation/v1 BandwithAllocationServiceServer > proto/bw_allocation/v1/mock_bw_allocation/mock.go

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,paths=source_relative proto/bwreserver/v1/bwreserver.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,paths=source_relative proto/bw_allocation/v1/bw_allocation.proto

test:
	go test ./...
