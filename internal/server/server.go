package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/dlinh31/go-videostream/proto"
	"google.golang.org/grpc"
)

type VideoStreamServer struct {
    pb.UnimplementedVideoStreamServiceServer
    VideoDir string
}

func RegisterServer(grpcServer *grpc.Server, videoDir string) {
    pb.RegisterVideoStreamServiceServer(grpcServer, &VideoStreamServer{VideoDir: videoDir})
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

var watchParties = make(map[string][]string)
var mu sync.Mutex



func (*VideoStreamServer) CreateWatchParty(ctx context.Context ,req *pb.CreatePartyRequest) (*pb.PartyResponse, error){
    mu.Lock()
    defer mu.Unlock()
    partyId := fmt.Sprintf("party-%d", len(watchParties) + 1)
    watchParties[partyId] = []string{req.HostName}
    return &pb.PartyResponse{PartyId: partyId, Status: "success",Users: watchParties[partyId]}, nil
}

func (*VideoStreamServer) JoinWatchParty(ctx context.Context, req *pb.JoinPartyRequest) (*pb.PartyResponse, error){
    mu.Lock()
    defer mu.Unlock()
    users, exists := watchParties[req.PartyId];
    if !exists{
        return &pb.PartyResponse{PartyId: req.PartyId, Status: "party not found", Users: nil}, nil
    }
    watchParties[req.PartyId] = append(users, req.UserName)
    return &pb.PartyResponse{PartyId: req.PartyId, Status: "success", Users: watchParties[req.PartyId]}, nil
}