package main

import (
	"log"
	"net"

	"github.com/dlinh31/go-videostream/internal/server"
	"google.golang.org/grpc"
)

const port = ":8080"

func main() {
    videoDir := "videos"
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    grpcServer := grpc.NewServer()
    server.RegisterServer(grpcServer, videoDir)

    log.Printf("Server started at %v", lis.Addr())
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
