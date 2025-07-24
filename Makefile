.PHONY: dockerpb build run clean test

proto:
	protoc --proto_path=pb --go_out=pb --go_opt=module=github.com/Daylily-kor/daylily-grpc-server/pb --go-grpc_out=pb --go-grpc_opt=module=github.com/Daylily-kor/daylily-grpc-server/pb pb/*.proto

build:
	go build -o bin/server.exe ./cmd/server

run: build
	./bin/server.exe

clean:
	rm -rf bin/

test:
	go test ./...