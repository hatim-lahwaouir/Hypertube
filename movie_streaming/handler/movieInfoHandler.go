package handler


import (
    "fmt"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/dto"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
    "strconv"
    "net/http"
)


type Movie struct  {
    MovieInfoService *services.DownloadMovieInfoService
}

func NewMovieHandler(m  *services.DownloadMovieInfoService) *Movie{
    return &Movie{MovieInfoService: m}
}



func (m *Movie) Hello(w http.ResponseWriter, r *http.Request) error{

    // download a specific movie
    imdb_code := r.PathValue("imdb_code")
    
    movie, err := m.MovieInfoService.DownloadMovieInfo(imdb_code)

    if err != nil {
        fmt.Println(err)
        return utils.WriteResp(w, http.StatusNotFound, err.Error())
    }
    return utils.WriteResp(w, http.StatusCreated, movie)
}





func (m *Movie) MovieSuggestions(w http.ResponseWriter, r *http.Request) error{
    

    page , err := strconv.ParseUint(r.PathValue("page"), 10, 64)
    if err != nil {
        return utils.WriteResp(w, http.StatusBadRequest, "Invalid count param")
    }




    moviesFilter,  field_errors := dto.NewMovieFilters(r.Body) 

    if moviesFilter == nil {
        return utils.WriteResp(w, http.StatusBadRequest, field_errors)

    }
    movies, err := m.MovieInfoService.SearchForMovies(moviesFilter, page)

    if err != nil {
        fmt.Println(err)
        return utils.WriteResp(w, http.StatusInternalServerError , "InternalServerError")
    }
    return utils.WriteResp(w, http.StatusOK, movies)
}




func (m *Movie) MovieSuggersionsOMDB(w http.ResponseWriter, r *http.Request) error{
    

    page , err := strconv.ParseUint(r.PathValue("page"), 10, 64)
    if err != nil {
        return utils.WriteResp(w, http.StatusBadRequest, "Invalid count param")
    }




    moviesFilter,  field_errors := dto.NewMovieFiltersOMDB(r.Body) 

    if moviesFilter == nil {
        return utils.WriteResp(w, http.StatusBadRequest, field_errors)

    }
    movies, err := m.MovieInfoService.SearchForMoviesOMDB(moviesFilter, page)
    if err != nil {
        return utils.WriteResp(w, http.StatusNotFound, err.Error())
    }
    return utils.WriteResp(w, http.StatusOK, movies)
}


