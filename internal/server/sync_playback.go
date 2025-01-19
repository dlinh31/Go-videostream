package server

import (
	"context"
	"io"
	"log"
	"net/http"

	pb "github.com/dlinh31/go-videostream/proto"
)

type SyncStream struct {
    PartyID string
    Stream  pb.VideoStreamService_SyncPlaybackServer
}
type PlaybackState struct {
    Timestamp int64
    IsPlaying bool
}
var playbackStreams = make(map[string][]SyncStream) // Map party_id to streams
var playbackStates = make(map[string]*PlaybackState)

func (s *VideoStreamServer) SyncPlayback(stream pb.VideoStreamService_SyncPlaybackServer) error {
    var partyID string
    defer func() {
        if partyID != "" {
            mu.Lock()
            for i, syncStream := range playbackStreams[partyID] {
                if syncStream.Stream == stream {
                    playbackStreams[partyID] = append(playbackStreams[partyID][:i], playbackStreams[partyID][i+1:]...)
                    break
                }
            }
            mu.Unlock()
        }
    }()

    for {
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
        if playbackStates[partyID] == nil {
			playbackStates[partyID] = &PlaybackState{}
		}
		if cmd.Command == "play" {
			playbackStates[partyID].IsPlaying = true
		} else if cmd.Command == "pause" {
			playbackStates[partyID].IsPlaying = false
		}
        playbackStates[partyID].Timestamp = cmd.Timestamp
		mu.Unlock()

        // Broadcast the command
        mu.Lock()
        for _, syncStream := range playbackStreams[partyID]{
            if syncStream.Stream != stream {
				log.Printf("Broadcasting command: %s to %s", cmd.Command, syncStream.PartyID)
                err := syncStream.Stream.Send(cmd)
                if err != nil {
					log.Printf("Error broadcasting to stream: %v", err)
				}
            }
        }
        mu.Unlock()

        // Add stream if not already present
        mu.Lock()
        streamExists := false
        for _, syncStream := range playbackStreams[partyID] {
            if syncStream.Stream == stream {
                streamExists = true
                break
            }
        }
        if !streamExists {
            playbackStreams[partyID] = append(playbackStreams[partyID], SyncStream{
                PartyID: partyID,
                Stream:  stream,
            })
        }
        mu.Unlock()
    }
    return nil
}


func (s *VideoStreamServer) SyncPlaybackWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Create a gRPC SyncPlayback stream
	stream, err := s.Client.SyncPlayback(context.Background())
	if err != nil {
		log.Printf("Error creating gRPC stream: %v", err)
		return
	}

	// Handle incoming WebSocket messages
	go func() {
		for {
			var cmd pb.PlaybackCommand
			err := conn.ReadJSON(&cmd)
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				stream.CloseSend()
				break
			}
			if err := stream.Send(&cmd); err != nil {
				log.Printf("Error sending gRPC command: %v", err)
				break
			}
		}
	}()

	// Relay gRPC responses back to WebSocket
	for {
		cmd, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving gRPC response: %v", err)
			break
		}
		if err := conn.WriteJSON(cmd); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}
