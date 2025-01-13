package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

    rangeHeader := r.Header.Get("Range")
    if rangeHeader == "" {
        w.Header().Set("Content-Type", "video/mp4")
        w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
        http.ServeContent(w, r, videoName, stat.ModTime(), file)
        return
    }

    var start, end int64
    _, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
    if err != nil || start >= fileSize {
        http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
        return
    }
    if end == 0 || end >= fileSize {
        end = fileSize - 1
    }

    contentLength := end - start + 1
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
}
