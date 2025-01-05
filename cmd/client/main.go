package main

import (
	"log"

	"github.com/dlinh31/go-videostream/internal/client"
)

func main() {
	client, err := client.NewClient("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// List available videos
	videos, err := client.ListVideos()
	if err != nil {
		log.Fatalf("Error while listing videos: %v", err)
	}
	log.Printf("Available videos: %v", videos)

	// Stream a specific video
	videoName := "sample2.mp4"
	if err := client.StreamVideo(videoName); err != nil {
		log.Fatalf("Error while streaming video: %v", err)
	}
}
