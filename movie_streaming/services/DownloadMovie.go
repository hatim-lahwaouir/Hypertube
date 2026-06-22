package services

import (
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/dto"
    "net/url"
    "compress/gzip"
    "path/filepath"
    "encoding/json"
    "strconv"
    "fmt"
    "net/http"
     "github.com/google/uuid"
    "time"
    "io"
    "os"
)

type DownloadMovieInfoService struct {
    Client *http.Client 
    TorrentPath string
    MoviesRep *models.MoviesRep
    MoviesToDownload chan types.Movie
}




var cache *DownloadMovieInfoService


func NewDownloadMovieInfoService(moviesRep *models.MoviesRep) *DownloadMovieInfoService {
    if  cache == nil {
        cache = &DownloadMovieInfoService{Client: &http.Client{}, TorrentPath: os.Getenv("TORRENT_PATH"), MoviesRep: moviesRep , MoviesToDownload: make(chan types.Movie, 100) }   
        cache.StartTask()
    }

    return cache 
}

func (m *DownloadMovieInfoService) SaveMoviesTask() {

    var  (
    
        torrents  []models.Torrent 
        movies []types.Movie
    )
    for movie := range m.MoviesToDownload {
        if filename, err := m.DownloadImg(movie.LargeCoverImage); err == nil {
            movie.LargeCoverImage= filename
        }
        // start downloading the torrants
        for _, val  := range movie.Torrents { 
            path, err := m.DownloadCompressedTorrantFiles(val)
            if err != nil {
                fmt.Println(">>>> error downloading torrent ", err)
                continue
            }
            torrents = append(torrents, models.Torrent{MovieID: movie.ID , Path: path, Size: val.Size, Quality : val.Quality, 
            Seeds: val.Seeds, Peers: val.Peers , Hash: val.Hash})
        }
        


        movies = append(movies, movie) 
        fmt.Println(">>>> download movie", movies)
        if len(movies) >= 5 {
            m.MoviesRep.CreateMovies(movies)
            m.MoviesRep.CreateTorrents(torrents)
            movies = movies[:0]
            torrents = torrents[:0]
        }
    }
}

func (m *DownloadMovieInfoService) StartTask() {
    go m.SaveMoviesTask()
}



func (m *DownloadMovieInfoService) DownloadMovieInfo(MovieId string) ([]models.Movie, error)  {


    // check if movie data already exists 
    movie, exists , err := m.MoviesRep.GetMovie(MovieId)
    if err != nil {
        return nil, fmt.Errorf("porblem connecting to database")
    }

    if exists {
        fmt.Println("movie exists")
        return []models.Movie{*movie}, nil 
    }
        var (
            movieResp types.MovieResponse
        )

        // parssing url and adding query to it 
        URL, _ := url.Parse("https://movies-api.accel.li/api/v2/movie_details.json")
        queries := URL.Query()
        queries.Add("imdb_id", MovieId)
        URL.RawQuery = queries.Encode()
        fmt.Println(URL.String())
        req, _ := http.NewRequest("GET", URL.String(), nil)

        resp, err := m.Client.Do(req)
        if err != nil {
            return nil, fmt.Errorf("failled to get movie info")
        }

        err = json.NewDecoder(resp.Body).Decode(&movieResp)
         if err != nil {
            return nil, fmt.Errorf("failled to get movie info")
        }
        if movieResp.Data.Movie.ID == 0 {
            return nil, fmt.Errorf("we couldn't find torrent of this movie")
        }

        m.MoviesToDownload <- movieResp.Data.Movie
        // save movie

    return m. YtsResponsToModels([]types.Movie{movieResp.Data.Movie}), nil
}




func (m *DownloadMovieInfoService) DownloadCompressedTorrantFiles(data types.Torrent) (string, error) {

    var (
        fileName string
        filePath string

    )

    req, err := http.NewRequest("GET", data.URL, nil)
    if err != nil {

        return "", fmt.Errorf("error doing the request to get the img")
    }
    
    resp, err := m.Client.Do(req)
    if err != nil {
        return "", fmt.Errorf("torrant URl isn't working")
    }
    defer resp.Body.Close()


    fileName =  fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
    filePath = filepath.Join(m.TorrentPath, fileName)


    dst , err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("error creating file for torrent")
    }
    defer dst.Close()
    gzipWriter, err := gzip.NewWriterLevel(dst, gzip.BestSpeed)
    if err != nil {
        return "", fmt.Errorf("error compressing torrent file %s", err.Error())
    }
    defer gzipWriter.Close()


    _ , err = io.Copy(gzipWriter, resp.Body)
    if err != nil {
        return "", fmt.Errorf("err in copy while downloading the torrent")
    }

    return fileName, nil
}

func (m *DownloadMovieInfoService) DownloadImg(URL string) (string, error) {

    var (
        fileName string
        filePath string

    )

    req, err := http.NewRequest("GET", URL, nil)
    if err != nil {

        return "", fmt.Errorf("error doing the request to get the img")
    }
    
    resp, err := m.Client.Do(req)
    if err != nil {
        return "", fmt.Errorf("img URl isn't working")
    }
    defer resp.Body.Close()


    fileName =  fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
    filePath = filepath.Join(m.TorrentPath, fileName)


    dst , err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("error creating file for torrent")
    }
    defer dst.Close()
    _ , err = io.Copy(dst, resp.Body)
    if err != nil {
        return "", fmt.Errorf("err in copy while downloading the img")
    }

    return fileName, nil
}


func (m *DownloadMovieInfoService) YtsResponsToModels(data []types.Movie) ([]models.Movie)  {
    var (
        movies []models.Movie
    )


    imdbGenres := types.GetGenres()
    for _, val := range(data) {
        var genres []models.Genre
        for _, valg := range val.Genres {
            genres = append(genres, models.Genre{Id: imdbGenres[valg], Type: valg})
        }
        m := models.Movie{Id: val.ID,
            Name: val.Title,
            IMDBCode: val.IMDBCode,
            UpdatedAt: time.Now(),
            Year: val.Year,
            Description: val.DescriptionFull,
            Rating: val.Rating,
            Genre : genres,
            Thumbnail: val.LargeCoverImage,
        }
        movies = append(movies, m)
    }
    return movies
}

func (m *DownloadMovieInfoService) SearchForMovies(data * dto.MovieFilters, pagination uint64) ([]models.Movie, error) {

     var (
            movieResp types.MoviesResponse
    //        torrents  []models.Torrent 
     )

    // fist search for movies inside the database
    movies, err := m.MoviesRep.GetMoviesWithFilters(data, pagination)
    // then let's call the api
    // here we need to get movie info to user
    // and start a thread that will install all these movies informations
    
    if len(movies) >= 10{
        fmt.Println("from db bitch")
        return movies, nil
    }

    URL, _ := url.Parse("https://movies-api.accel.li/api/v2/list_movies.json")
    queries := URL.Query()
    if data.Genre != "" {
        queries.Add("genre", data.Genre)
    }
    if data.Name != "" {
        queries.Add("query_term", data.Name)
    }
    if data.OrderBy != "" {
        queries.Add("order_by", data.OrderBy)
    }
    if data.SortBy!= "" {
        queries.Add("sort_by", data.SortBy)
    }


    queries.Add("limit", "10")
    queries.Add("page", strconv.FormatUint(pagination, 10))

    URL.RawQuery = queries.Encode()
    req, _ := http.NewRequest("GET", URL.String(), nil)

    resp, err := m.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failled to get movie info")
    }

    err = json.NewDecoder(resp.Body).Decode(&movieResp)
    if err != nil {
        return nil, fmt.Errorf("failled to get movie info")
    }


    // send movies info to be downloaded
    for  _, movie  :=  range movieResp.Data.Movie {
        m.MoviesToDownload <- movie
    }



    return  m.YtsResponsToModels(movieResp.Data.Movie), nil  
}



func (m *DownloadMovieInfoService) SearchForMoviesOMDB(data * dto.MovieFiltersOMDB, page uint64) (*types.MovieSearchResponseOMDB, error) {

    var (
            movieResp types.MovieSearchResponseOMDB
    )

    URL, _ := url.Parse("http://www.omdbapi.com/")
    queries := URL.Query()
    
    queries.Add("apikey", os.Getenv("OMDB_API_KEY"))
    queries.Add("s", data.Name)
    queries.Add("type", "movie")
    queries.Add("page", strconv.FormatUint(page, 10))
    if data.Year != "" {
        queries.Add("y", data.Year)
    }
    URL.RawQuery = queries.Encode()

    req, _ := http.NewRequest("GET", URL.String(), nil)
    resp, err := m.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failled to get more movies info")
    }

    err = json.NewDecoder(resp.Body).Decode(&movieResp)
    if err != nil {
        return nil, fmt.Errorf("failled to get more  movies info")
    }

    if movieResp.TotalResults == "" {
        return nil, fmt.Errorf("be more specified, too many results")
    }
    return &movieResp, nil
}
