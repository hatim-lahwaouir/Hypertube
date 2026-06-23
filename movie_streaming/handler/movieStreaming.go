package handler


import (
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
    "strconv"
    "fmt"
    "net/http"
)




type MovieStreaming struct  {
    MoviStreamingService *services.MovieStreamingService
    MovieRep             *models.MoviesRep
}


func NewMovieStreamingHandler(ms *services.MovieStreamingService, mrep *models.MoviesRep) *MovieStreaming{
    return &MovieStreaming{MoviStreamingService : ms,MovieRep: mrep}
}

func (ms *MovieStreaming) Download(w http.ResponseWriter, r *http.Request) error{
    movie_id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
    hash := r.PathValue("hash")
    if err != nil {
        return utils.WriteResp(w, http.StatusBadRequest, "Invalid movie id param")
    }

    t, err := ms.MovieRep.GetTorrent(movie_id, hash)
    if  err != nil {
        fmt.Println(">>>>>", err.Error())
        return utils.WriteResp(w, http.StatusBadRequest, "Invalid movie id param")
    }
    err = ms.MoviStreamingService.ParseTorrent(t)
    if err != nil {
        fmt.Println(t, err.Error())
    }


    return utils.WriteResp(w, http.StatusCreated, t)
}
