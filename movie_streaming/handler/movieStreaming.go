package handler

import (
	"fmt"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
	"net/http"
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
	err = ms.MoviStreamingService.ParseTorrent(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	return utils.WriteResp(w, http.StatusCreated, "good")
}
