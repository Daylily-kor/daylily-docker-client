package main

import (
	"context"
	"flag"
	"log"
	"net"

	greeterProto "github.com/Daylily-kor/daylily-docker-client/proto/greeter"
	"google.golang.org/grpc"
)

type server struct {
	greeterProto.UnimplementedGreeterServer
}

func (s *server) SayHello(_ context.Context, in *greeterProto.HelloReq) (*greeterProto.HelloResp, error) {
	log.Printf("Received: %v", in.GetName())
	return &greeterProto.HelloResp{
		Message: "Hello from Go",
		Code:    69,
		Error:   "",
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	greeterProto.RegisterGreeterServer(s, &server{})
	log.Printf("Server is listening on port :50051")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
