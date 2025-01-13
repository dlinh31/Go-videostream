package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/dlinh31/go-videostream/proto"
)


func (s *VideoStreamServer) ListVideos(ctx context.Context, req *pb.NoParam) (*pb.VideoList, error) {
    files, err := os.ReadDir(s.VideoDir)
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
    videoPath := filepath.Join(s.VideoDir, req.VideoName)
    file, err := os.Open(videoPath)
    if err != nil {
        return fmt.Errorf("cannot open video file: %w", err)
    }
    defer file.Close()

    const chunkSize = 1024 * 1024 // 1 MB
    buffer := make([]byte, chunkSize)
    for {
        n, err := file.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("error while reading video file: %w", err)
        }

        if err := stream.Send(&pb.VideoChunk{ChunkData: buffer[:n]}); err != nil {
            return fmt.Errorf("error while sending video chunk: %w", err)
        }

    }
    return nil
}