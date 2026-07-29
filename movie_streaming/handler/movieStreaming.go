package handler

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
)

type MovieStreaming struct {
	MoviStreamingService *services.MovieStreamingService

}

func NewMovieStreamingHandler(ms *services.MovieStreamingService) *MovieStreaming {
	return &MovieStreaming{MoviStreamingService: ms,}
}

func (ms *MovieStreaming) Download(w http.ResponseWriter, r *http.Request) error {
    r.ParseMultipartForm(20 << 20)

    file, _, err := r.FormFile("torrent")
    if err != nil {
		return utils.WriteResp(w, http.StatusBadRequest, "Invalid torrent name")
    }	
	defer file.Close()
	infoHash, err := ms.MoviStreamingService.ParseTorrent(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	return utils.WriteResp(w, http.StatusCreated, infoHash)
}



func (ms *MovieStreaming) StreamVideo(w http.ResponseWriter, r *http.Request) error {

	infoHash,err := hex.DecodeString(r.PathValue("infohash"))
	
	if err != nil {
		return utils.WriteResp(w, http.StatusBadRequest, "Invalid Info hash")
	}
	rangeHeader := r.Header.Get("Range")
    var start, end int64
    start = 0
    end = 0

    if rangeHeader != "" {
        rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
        parts := strings.Split(rangeStr, "-")
        if len(parts) > 0 && parts[0] != "" {
            start, _ = strconv.ParseInt(parts[0], 10, 64)
        }
        if len(parts) > 1 && parts[1] != "" {
            end, _ = strconv.ParseInt(parts[1], 10, 64)
        }
    }
	if end == 0 {
		end = start + 5000000 - 1
	}


	
	fmt.Println("start", start, rangeHeader)
	fmt.Println("end", end)
	data, err := ms.MoviStreamingService.HasBitField(infoHash, uint64(start), uint64(end))

	if err != nil {
		return utils.WriteResp(w, http.StatusBadRequest, err.Error())
	}
	w.Header().Set("Content-Type", "video/mp4") // Consider mapping this dynamically based on file extension
    w.Header().Set("Accept-Ranges", "bytes")
    w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
    w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start + int64(len(data)) - 1, ms.MoviStreamingService.FileSize(infoHash)))

	fmt.Println("------------->>>>>>>>>>>>>", len(data), start, end)
	w.WriteHeader(http.StatusPartialContent)
	_, err = w.Write(data)
    if err != nil {
        // If w.Write fails (e.g., client closed the browser), do NOT return the error 
        // back to MakeHandler. If you do, MakeHandler will try to write a 500 JSON error 
        // over the video stream, causing the "superfluous WriteHeader" panic.
        fmt.Println("Client disconnected or stream interrupted:", err)
        return nil 
    }

    // 3. Return nil so MakeHandler knows the request succeeded
    return nil
	// return utils.WriteResp(w, http.StatusPartialContent, data)
}