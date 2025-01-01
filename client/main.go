package main

import (
	"context"
	"io"
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

	videoName := "sample2.mp4"
	streamVideo(client, videoName)

}

func streamVideo(client pb.VideoStreamServiceClient, videoName string){
	req := &pb.VideoRequest{VideoName: videoName}
	stream, err := client.StreamVideo(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to stream video: %v", err)
	}
	log.Printf("Start streaming video: %s", videoName)
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Println("Video streaming completed")
			break
		}
		if err != nil {
			log.Fatalf("Error while receiving video chunk: %v", err)
		}
		log.Printf("Received chunk of size: %d bytes", len(chunk.ChunkData))
	}

}