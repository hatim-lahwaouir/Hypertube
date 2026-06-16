package handler


import (
   // "fmt"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/dto"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
    "net/http"
    "fmt"
)


type Movie struct  {
    MovieService *services.DownloadMovieService
}

func NewMovieHandler(m  *services.DownloadMovieService) *Movie{
    return &Movie{MovieService: m}
}



func (m *Movie) Hello(w http.ResponseWriter, r *http.Request) error{

    // download a specific movie
    

    m.MovieService.DownloadMovieInfo(794)
    return utils.WriteResp(w, http.StatusCreated, "Account created wait for an email will be sent to you ! ")
}





func (m *Movie) MovieSuggestions(w http.ResponseWriter, r *http.Request) error{
    

    moviesFilter,  field_errors := dto.NewMovieFilters(r.Body) 

    if moviesFilter == nil {
        return utils.WriteResp(w, http.StatusBadRequest, field_errors)

    }
    m.MovieService.SearchForMovies(moviesFilter)
    fmt.Println(moviesFilter)
    return utils.WriteResp(w, http.StatusCreated, "Account created wait for an email will be sent to you ! ")
}





