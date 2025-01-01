package main

import (
	"context"
	"log"
	"time"

	pb "github.com/dlinh31/go-videostream/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:8080"
)

func main(){
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("client cannot connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewVideoStreamServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(),time.Second * 5)
	defer cancel()
	res, err := client.ListVideos(ctx, &pb.NoParam{})
	if err != nil {
		log.Fatalf("Error when calling ListVideos: %v", err)
	}
	log.Printf("Available videos: %v", res.Videos)

}