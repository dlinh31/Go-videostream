package client

import (
	"context"
	"io"
	"log"
	"time"

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
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewVideoStreamServiceClient(conn)
	return &VideoClient{conn: conn, client: client}, nil
}

// Close closes the client connection
func (vc *VideoClient) Close() {
	vc.conn.Close()
}

// ListVideos retrieves a list of available videos from the server
func (vc *VideoClient) ListVideos() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := vc.client.ListVideos(ctx, &pb.NoParam{})
	if err != nil {
		return nil, err
	}
	return res.Videos, nil
}

// StreamVideo streams a specific video from the server
func (vc *VideoClient) StreamVideo(videoName string) error {
	req := &pb.VideoRequest{VideoName: videoName}
	stream, err := vc.client.StreamVideo(context.Background(), req)
	if err != nil {
		return err
	}

	log.Printf("Start streaming video: %s", videoName)
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Println("Video streaming completed")
			break
		}
		if err != nil {
			return err
		}
		log.Printf("Received chunk of size: %d bytes", len(chunk.ChunkData))
	}
	return nil
}

func (vc *VideoClient) CreateWatchParty(hostName string) (string, error){
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	res, err := vc.client.CreateWatchParty(ctx, &pb.CreatePartyRequest{HostName: hostName})
	if err != nil {
		return "", err
	}
	log.Printf("Created watch party: %s, Participants: %v", res.PartyId, res.Users)
	return res.PartyId, nil
}

func (vc *VideoClient) JoinWatchParty(partyId, userName string) (error){
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	res, err := vc.client.JoinWatchParty(ctx, &pb.JoinPartyRequest{PartyId: partyId, UserName: userName})
	if err != nil {
		return err
	}
	log.Printf("Joined watch party: %s, Participants: %v", res.PartyId, res.Users)
	return nil
}