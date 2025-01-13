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

    partyID := "party-1"
    userName := "HostUser"

    err = client.SyncPlayback(partyID, userName)
    if err != nil {
        log.Fatalf("Error during playback sync: %v", err)
    }

	
}
