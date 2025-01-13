package client

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/dlinh31/go-videostream/proto"
)
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
func (vc *VideoClient) StreamAndSaveVideo(videoName string) (string, error) {
	req := &pb.VideoRequest{VideoName: videoName}
	stream, err := vc.client.StreamVideo(context.Background(), req)
	if err != nil {
		return "", err
	}

	// Create a temporary file for playback
	tempFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	log.Printf("Streaming video: %s", videoName)

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Println("Video streaming completed")
			break
		}
		if err != nil {
			return "", err
		}

		// Write chunk data to the temporary file
		if _, err := tempFile.Write(chunk.ChunkData); err != nil {
			return "", err
		}
	}

	return tempFile.Name(), nil
}
