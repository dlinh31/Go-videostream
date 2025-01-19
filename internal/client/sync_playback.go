package client

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/dlinh31/go-videostream/proto"
)

func (vc *VideoClient) SyncPlayback(partyID, userName string) error {
	// Create a bidirectional stream
	stream, err := vc.client.SyncPlayback(context.Background())
	if err != nil {
		return err
	}
	log.Println("Connected to SyncPlayback stream")

	// Goroutine to send commands
	go func() {
		commands := []struct {
			Command   string
			Timestamp int64
		}{
			{"play", 0},
			{"pause", 5},
			{"play", 10},
		}

		for _, cmd := range commands {
			time.Sleep(3 * time.Second) // Delay between commands
			log.Printf("Sending %s command at timestamp %d", cmd.Command, cmd.Timestamp)
			err := stream.Send(&pb.PlaybackCommand{
				PartyId:   partyID,
				Command:   cmd.Command,
				Timestamp: cmd.Timestamp,
				UserName:  userName,
			})
			if err != nil {
				log.Printf("Error sending %s command: %v", cmd.Command, err)
				return
			}
		}

		// Close the sending stream once commands are sent
		if err := stream.CloseSend(); err != nil {
			log.Printf("Error closing send stream: %v", err)
		}
	}()

	// Main loop to receive and process commands
	for {
		cmd, err := stream.Recv()
		if err == io.EOF {
			log.Println("Server closed the stream")
			break
		}
		if err != nil {
			log.Printf("Error receiving playback command: %v", err)
			return err
		}

		// Log received commands for debugging
		log.Printf("Received command: %s at %d from %s", cmd.Command, cmd.Timestamp, cmd.UserName)
	}
	return nil
}
