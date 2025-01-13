package server

import (
	"io"
	"log"

	pb "github.com/dlinh31/go-videostream/proto"
)

type SyncStream struct {
    PartyID string
    Stream  pb.VideoStreamService_SyncPlaybackServer
}
var playbackStreams = make(map[string][]SyncStream) // Map party_id to streams

func (s *VideoStreamServer) SyncPlayback(stream pb.VideoStreamService_SyncPlaybackServer) error {
    var partyID string
    for {
        // Receive playback commands
        cmd, err := stream.Recv()
        if err != nil {
            if err == io.EOF {
                log.Println("Stream closed")
                break
            }
            log.Printf("Error receiving playback command: %v", err)
            return err
        }

        log.Printf("Received command: %s at %d from %s", cmd.Command, cmd.Timestamp, cmd.UserName)
        partyID = cmd.PartyId

        // Broadcast the command
        mu.Lock()
        for _, syncStream := range playbackStreams[partyID] {
            if syncStream.Stream != stream {
                log.Printf("Broadcasting command: %s to %s", cmd.Command, syncStream.PartyID)
                syncStream.Stream.Send(cmd)
            }
        }
        mu.Unlock()

        // Add stream if not already present
        mu.Lock()
        playbackStreams[partyID] = append(playbackStreams[partyID], SyncStream{
            PartyID: partyID,
            Stream:  stream,
        })
        mu.Unlock()
    }
    return nil
}
