package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/dlinh31/go-videostream/proto"
	"google.golang.org/grpc"
)

const(
	port = ":8080"
)
type VideoStreamServer struct{
	pb.VideoStreamServiceServer
	videoDir string
}

func (s *VideoStreamServer) ListVideos(ctx context.Context, req *pb.NoParam) (*pb.VideoList, error){
	files, err := os.ReadDir(s.videoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read video directory: %w", err)
	}
	
	var videos []string
	for _, file := range files {
		if !file.IsDir() {
			videos = append(videos, file.Name())	
		}
	}
	return &pb.VideoList{Videos: videos}, nil
}

// func (s *VideoStreamServer) StreamVideo (ctx context.Context, req *pb.VideoRequest, stream pb.VideoStreamService_StreamVideoServer) error {
// 	return nil
// }


func main(){
	videoDir := "videos"
	lis, err := net.Listen("tcp", port)
	if err != nil{
		log.Fatalf("Failed to start server %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterVideoStreamServiceServer(grpcServer, &VideoStreamServer{videoDir: videoDir})
	log.Printf("server started at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err!= nil {
		log.Fatalf("Failed to start server %v", err)
	}
	if err := grpcServer.Serve(lis); err!= nil {
		log.Fatalf("Failed to start server %v", err)
	}
}