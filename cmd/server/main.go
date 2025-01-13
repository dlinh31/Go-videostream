package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dlinh31/go-videostream/internal/server"
	"google.golang.org/grpc"
)

const (
	grpcPort = "localhost:50051" // gRPC runs on a different port
	httpPort = "localhost:8080"  // REST API port
)

func main() {
	// Resolve the absolute path for the video directory
	videoDir, err := filepath.Abs("videos")
	if err != nil {
		log.Fatalf("Failed to resolve video directory: %v", err)
	}

	// Verify that the video directory exists and is accessible
	if _, err := os.Stat(videoDir); os.IsNotExist(err) {
		log.Fatalf("Video directory does not exist: %s", videoDir)
	}
	log.Printf("Video directory resolved to: %s", videoDir)

	// Start the REST server in a separate goroutine
	go func() {
		mux := http.NewServeMux()

		// Register REST handlers with logging
		mux.HandleFunc("/api/videos", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request received at /api/videos")
			server.ListVideosHandler(w, r)
		})

		mux.HandleFunc("/api/video", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request received at /api/video")
			server.ServeVideoHandler(w, r)
		})

		mux.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request received at /api/stream")
			server.StreamVideoHandler(w, r)
		})

		// Apply CORS middleware
		corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")
			w.Header().Set("Accept-Ranges", "bytes")
	
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			mux.ServeHTTP(w, r)
		})

		log.Printf("Starting REST server on %s", httpPort)
		if err := http.ListenAndServe(httpPort, corsMux); err != nil {
			log.Fatalf("Failed to start REST server: %v", err)
		}
	}()

	// Start the gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to start gRPC listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	server.RegisterServer(grpcServer, videoDir)

	log.Printf("Starting gRPC server on %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
