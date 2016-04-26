package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	"com.grid/chsen/gsio/queue"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	queue.RegisterQueueServiceServer(server, queue.NewQueue())
	server.Serve(lis)
}