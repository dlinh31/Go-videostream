package server

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
    pb "github.com/dlinh31/go-videostream/proto"
    "google.golang.org/grpc"
)

type VideoStreamServer struct {
    pb.UnimplementedVideoStreamServiceServer
    VideoDir string
}

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

        time.Sleep(500 * time.Millisecond) // Simulate delay
    }
    return nil
}

func RegisterServer(grpcServer *grpc.Server, videoDir string) {
    pb.RegisterVideoStreamServiceServer(grpcServer, &VideoStreamServer{VideoDir: videoDir})
}
