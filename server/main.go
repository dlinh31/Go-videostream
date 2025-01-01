package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

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

func (s *VideoStreamServer) StreamVideo(req *pb.VideoRequest, stream pb.VideoStreamService_StreamVideoServer) error {
	videoPath := filepath.Join(s.videoDir, req.VideoName)
	file, err := os.Open(videoPath)
	if err != nil {
		log.Fatalf("Cannot open file with given directory in server %v", err)
	}
	defer file.Close()
	chunkSize := 1024 * 2024 // 1MB
	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF{
			break
		}
		if err != nil {
			log.Fatalf("Error while read video in server %v", err)
			return err
		}
		if err := stream.Send(&pb.VideoChunk{ChunkData: buffer[:n]}); err != nil {
			log.Fatalf("Error while sending video chunk in server %v", err)
			return err
		}
		time.Sleep(500 * time.Millisecond) // 500ms delay

	}
	return nil

}


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