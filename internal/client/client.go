package client

import (
	pb "github.com/dlinh31/go-videostream/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VideoClient struct {
	conn   *grpc.ClientConn
	client pb.VideoStreamServiceClient
}

// NewClient creates and returns a new VideoClient
func NewClient(serverAddress string) (*VideoClient, error) {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewVideoStreamServiceClient(conn)
	return &VideoClient{conn: conn, client: client}, nil
}

// Close closes the client connection
