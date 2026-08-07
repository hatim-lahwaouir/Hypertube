package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	// "time"

	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
)

type MovieStreaming struct {
	MoviStreamingService *services.MovieStreamingService

}

func NewMovieStreamingHandler(ms *services.MovieStreamingService) *MovieStreaming {
	return &MovieStreaming{MoviStreamingService: ms,}
}

type MagnetRequest struct {
    MagnetLink string `json:"magnet_link"`
}



func (ms *MovieStreaming) Download(w http.ResponseWriter, r *http.Request) error {

    var req MagnetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return utils.WriteResp(w, http.StatusBadRequest, "Invalid request body")
    }

    if req.MagnetLink == "" {
        return utils.WriteResp(w, http.StatusBadRequest, "magnet_link is required")
    }


    t := services.NewDownloadTorrent(req.MagnetLink)
	
	file, err :=  t.DownloadTorrent()
	if err != nil  {
		fmt.Println("err internal server error", err)
		return utils.WriteResp(w, http.StatusInternalServerError, "StatusInternalServerError")
	}

	defer file.Close()



	UUID , err := ms.MoviStreamingService.ParseTorrent(file)
	if err != nil {
		fmt.Println(err.Error())
		return utils.WriteResp(w, http.StatusInternalServerError, "StatusInternalServerError")
	}

	return utils.WriteResp(w, http.StatusCreated, UUID)
}



func (ms *MovieStreaming) StreamVideo(w http.ResponseWriter, r *http.Request) error {
	infoHash := r.PathValue("movie_id")
	isDownload := r.URL.Query().Get("action") == "download"
	totalFileSize := int64(ms.MoviStreamingService.FileSize(infoHash))

	rangeHeader := r.Header.Get("Range")
	var start, end int64 = 0, totalFileSize - 1

	// Parse the requested start byte if provided
	if rangeHeader != "" {
		rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
		parts := strings.Split(rangeStr, "-")
		if len(parts) > 0 && parts[0] != "" {
			start, _ = strconv.ParseInt(parts[0], 10, 64)
		}
		// We ignore the requested 'end' for downloads to force the whole file,
		// but respect it for video players if they specifically ask for a small chunk.
		if len(parts) > 1 && parts[1] != "" && !isDownload {
			end, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}

	// 1. FOR VIDEO PLAYERS: Cap the chunk size to save memory (e.g., 2MB chunks)
	if !isDownload {
		if end-start > 2000000 {
			end = start + 2000000 - 1
		}
		
		data, err := ms.MoviStreamingService.HasBitField(infoHash, uint64(start), uint64(end))
		ms.MoviStreamingService.UpdateTime(infoHash)
		if err != nil {
			return utils.WriteResp(w, http.StatusBadRequest, err.Error())
		}

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start+int64(len(data))-1, totalFileSize))
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"",ms.MoviStreamingService.Filename(infoHash)))
		
		w.WriteHeader(http.StatusPartialContent)
		w.Write(data)
		return nil
	}

	// 2. FOR DIRECT DOWNLOADS: Stream the entire file continuously
	// We use a loop to fetch chunks from your torrent engine so we don't load a 2GB movie into RAM all at once.
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(totalFileSize-start, 10))
	w.Header().Set("Content-Disposition",fmt.Sprintf("attachment; filename=\"%s\"",ms.MoviStreamingService.Filename(infoHash)))

	if start > 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, totalFileSize-1, totalFileSize))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	currentOffset := start
	chunkSize := int64(2000000) // 2MB internal chunks

	for currentOffset < totalFileSize {
		chunkEnd := currentOffset + chunkSize - 1
		if chunkEnd >= totalFileSize {
			chunkEnd = totalFileSize - 1
		}

		// This will block until the torrent engine has this specific 2MB piece
		data, err := ms.MoviStreamingService.HasBitField(infoHash, uint64(currentOffset), uint64(chunkEnd))
		ms.MoviStreamingService.UpdateTime(infoHash)
		if err != nil {
			fmt.Println("Error fetching piece during download:", err)
			return err
		}

		// Write the chunk to the HTTP response
		_, err = w.Write(data)
		if err != nil {
			fmt.Println("Client disconnected or download cancelled:", err)
			return nil // Client closed the connection, just exit cleanly
		}

		// Flush the HTTP writer to ensure data goes to the browser immediately
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		currentOffset = chunkEnd + 1
	}

	return nil
}



func (ms *MovieStreaming) StreamStatus(w http.ResponseWriter, r *http.Request) error {

	id := r.PathValue("movie_id")

	status := ms.MoviStreamingService.StreamStatus(id)
	
	return utils.WriteResp(w, http.StatusOK, status)
}

