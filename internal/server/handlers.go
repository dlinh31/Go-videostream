package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)


type Video struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func ListVideosHandler(w http.ResponseWriter, r *http.Request){
	videoDir := "videos"
	files, err := os.ReadDir(videoDir)
	if err != nil {
		http.Error(w, "Failed to read videos directory", http.StatusInternalServerError)
		return
	}
	var videos []Video

	for _, file := range files {
		if !file.IsDir(){
			videos = append(videos, Video{
				Name: file.Name(),
				Path: filepath.Join(videoDir, file.Name()),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func ServeVideoHandler(w http.ResponseWriter, r *http.Request){
	videoName := r.URL.Query().Get("name")
	if videoName == ""{
		http.Error(w, "Missing video name", http.StatusBadRequest)
		return
	}
	videoPath := filepath.Join("videos", videoName)
	http.ServeFile(w, r, videoPath)
}


func StreamVideoHandler(w http.ResponseWriter, r *http.Request) {
    videoName := r.URL.Query().Get("name")
    if videoName == "" {
        http.Error(w, "Missing video name", http.StatusBadRequest)
        return
    }

    videoPath := filepath.Join("videos", videoName)
    file, err := os.Open(videoPath)
    if err != nil {
        log.Printf("Failed to open video file: %v", err)
        http.Error(w, "Video not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        log.Printf("Failed to stat video file: %v", err)
        http.Error(w, "Unable to read video file", http.StatusInternalServerError)
        return
    }
    fileSize := stat.Size()
    log.Printf("File size: %d bytes", fileSize)

    rangeHeader := r.Header.Get("Range")
    log.Printf("Range header: %s", rangeHeader)

    var start, end int64
    if rangeHeader == "" {
        // No Range header, serve the full file
        start = 0
        end = fileSize - 1
    } else {
        // Parse the Range header
        _, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
        if err != nil {
            // Handle case where only the start is specified, e.g., "bytes=0-"
            _, err = fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
            if err != nil {
                log.Printf("Error parsing Range header: %v", err)
                http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
                return
            }
            end = fileSize - 1
        }

        // Validate the range
        if start < 0 || start >= fileSize || end >= fileSize || start > end {
            log.Printf("Invalid Range header: %s", rangeHeader)
            http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
            return
        }
    }

    contentLength := end - start + 1
    log.Printf("Serving byte range %d-%d (%d bytes)", start, end, contentLength)

    w.Header().Set("Content-Type", "video/mp4")
    w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
    w.Header().Set("Accept-Ranges", "bytes")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
    w.WriteHeader(http.StatusPartialContent)

    buffer := make([]byte, 1024*1024) // 1MB buffer
    file.Seek(start, io.SeekStart)
    for contentLength > 0 {
        bytesToRead := int64(len(buffer))
        if contentLength < bytesToRead {
            bytesToRead = contentLength
        }

        n, err := file.Read(buffer[:bytesToRead])
        if err != nil && err != io.EOF {
            log.Printf("Error reading video file: %v", err)
            break
        }

        w.Write(buffer[:n])
        contentLength -= int64(n)
    }
    log.Println("Finished streaming video.")
}


type WebSocketClient struct {
	Conn     *websocket.Conn
	PartyID  string
	UserName string
}

var (
	webSocketClients = make(map[string][]*WebSocketClient) // Map party_id to clients
	webSocketMu      sync.Mutex
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		return true
	},
}

func SyncPlaybackWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	partyID := r.URL.Query().Get("party_id")
	userName := r.URL.Query().Get("user_name")
	if partyID == "" || userName == "" {
		http.Error(w, "Missing party_id or user_name", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	client := &WebSocketClient{
		Conn:     conn,
		PartyID:  partyID,
		UserName: userName,
	}

	// Add the client to the party
	webSocketMu.Lock()
	webSocketClients[partyID] = append(webSocketClients[partyID], client)
	webSocketMu.Unlock()

	log.Printf("New WebSocket client connected: %s in party %s", userName, partyID)

	// Handle incoming messages from the client
	for {
		var msg struct {
			Command   string `json:"command"`
			Timestamp int64  `json:"timestamp"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			break
		}

		log.Printf("Received command: %s at %d from %s", msg.Command, msg.Timestamp, userName)

		// Broadcast the message to all clients in the party
		webSocketMu.Lock()
		for _, c := range webSocketClients[partyID] {
			if c != client {
				err := c.Conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error broadcasting to WebSocket client: %v", err)
				}
			}
		}
		webSocketMu.Unlock()
	}

	// Remove the client from the party on disconnect
	webSocketMu.Lock()
	for i, c := range webSocketClients[partyID] {
		if c == client {
			webSocketClients[partyID] = append(webSocketClients[partyID][:i], webSocketClients[partyID][i+1:]...)
			break
		}
	}
	webSocketMu.Unlock()
	log.Printf("WebSocket client disconnected: %s from party %s", userName, partyID)
}