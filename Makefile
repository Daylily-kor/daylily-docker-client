.PHONY: dockerpb build run clean test

dockerpb:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/dockerpb/dockerpb.proto

build:
	go build -o bin/server.exe ./cmd/server

run: build
	./bin/server.exe

clean:
	rm -rf bin/

test:
	go test ./...