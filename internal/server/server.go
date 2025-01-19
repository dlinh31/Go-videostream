package server

import (
	pb "github.com/dlinh31/go-videostream/proto"
	"google.golang.org/grpc"
)

type VideoStreamServer struct {
    pb.UnimplementedVideoStreamServiceServer
    VideoDir string
    Client   pb.VideoStreamServiceClient 
}
func RegisterServer(grpcServer *grpc.Server, videoDir string) {
    pb.RegisterVideoStreamServiceServer(grpcServer, &VideoStreamServer{VideoDir: videoDir})
}
