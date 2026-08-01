package handler

import (
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

var magnetLink string = "magnet:?xt=urn:btih:37E77490BC4F285DBFA837514715A20BD405A502&dn=Spider-Man%3A+Far+from+Home+%282019%29+%5BWEBRip%5D+%5B1080p%5D+%5BYTS%5D+%5BYIFY%5D&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2F9.rarbg.com%3A2710%2Fannounce&tr=udp%3A%2F%2Fp4p.arenabg.com%3A1337&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce"




func (ms *MovieStreaming) Download(w http.ResponseWriter, r *http.Request) error {
    // r.ParseMultipartForm(20 << 20)

    // file, _, err := r.FormFile("torrent")
    // if err != nil {
	// 	return utils.WriteResp(w, http.StatusBadRequest, "Invalid torrent name")
    // }	
	// defer file.Close()
	// install the torrent the passit to the parseTorrent

	t := services.NewDownloadTorrent(magnetLink)
	
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

	infoHash := r.PathValue("infohash")
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
		end = start + 2000000 - 1
	}


	
	fmt.Println("start", start, rangeHeader)
	fmt.Println("end", end)
	data, err := ms.MoviStreamingService.HasBitField(infoHash, uint64(start), uint64(end))
	ms.MoviStreamingService.UpdateTime(infoHash)
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

