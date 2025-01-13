package client

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/dlinh31/go-videostream/proto"
)

func (vc *VideoClient) SyncPlayback(partyID, userName string) error {
	stream, err := vc.client.SyncPlayback(context.Background())
	if err != nil {
		return err
	}
	log.Println("Starting sync playback")

	go func() {
		log.Println("Sending play command")
		err := stream.Send(&pb.PlaybackCommand{
			PartyId:   partyID,
            Command:   "play",
            Timestamp: 0, // Start from the beginning
            UserName:  userName,
		})
		if err != nil {
			log.Printf("Error sending playback command %v", err)
		}
	}()
	go func() {
		time.Sleep(time.Second * 3)
		log.Println("Sending play command")
		err := stream.Send(&pb.PlaybackCommand{
			PartyId:   partyID,
            Command:   "play",
            Timestamp: 3, // Start from the beginning
            UserName:  userName,
		})
		if err != nil {
			log.Printf("Error sending playback command %v", err)
		}
	}()

	for {
        cmd, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        log.Printf("Received command: %s at %d from %s", cmd.Command, cmd.Timestamp, cmd.UserName)
    }
	return nil
}