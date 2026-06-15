package services

import (
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
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

type DownloadMovieService struct {
    Client *http.Client 
    TorrentPath string
    MoviesRep *models.MoviesRep
}




var cache *DownloadMovieService


func NewDownloadMovieService(moviesRep *models.MoviesRep) *DownloadMovieService {
    if  cache == nil {
        cache = &DownloadMovieService{Client: &http.Client{}, TorrentPath: os.Getenv("TORRENT_PATH"), MoviesRep: moviesRep }   
    }

    return cache 
}



func (m *DownloadMovieService) DownloadMovieInfo(MovieId uint64) error  {


    // check if movie data already exists 
    exists, err := m.MoviesRep.MovieExists(MovieId)
    if err != nil {
        return fmt.Errorf("porblem connecting to database")
    }

    if exists {
        fmt.Println("movie exists")
        return nil 
    } else {
        var (
            movieResp types.MovieResponse
            torrents  []models.Torrent 
        )

        // parssing url and adding query to it 
        URL, _ := url.Parse("https://movies-api.accel.li/api/v2/movie_details.json")
        queries := URL.Query()
        queries.Add("movie_id", strconv.FormatUint(MovieId, 10))
        URL.RawQuery = queries.Encode()
        req, _ := http.NewRequest("GET", URL.String(), nil)

        resp, err := m.Client.Do(req)
        if err != nil {
            return fmt.Errorf("failled to get movie info")
        }

        err = json.NewDecoder(resp.Body).Decode(&movieResp)
         if err != nil {
            return fmt.Errorf("failled to get movie info")
        }
       
        // save movie


        if filename, err := m.DownloadImg(movieResp.Data.Movie.LargeCoverImage); err == nil {
            movieResp.Data.Movie.LargeCoverImage= filename
        }
        m.MoviesRep.CreateMovie(&movieResp.Data.Movie)
        // start downloading the torrants
        for _, val  := range movieResp.Data.Movie.Torrents { 
            path, err := m.DownloadCompressedTorrantFiles(val)
            if err != nil {
                fmt.Println(">>>> error downloading torrent ", err)
                continue
            }
            torrents = append(torrents, models.Torrent{MovieID: MovieId, Path: path, Size: val.Size, Quality : val.Quality, 
            Seeds: val.Seeds, Peers: val.Peers })
        }
        // save torrents to db 
        m.MoviesRep.CreateTorrents(torrents)
    }
    return nil
}




func (m *DownloadMovieService) DownloadCompressedTorrantFiles(data types.Torrent) (string, error) {

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
    gzipWriter, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
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

func (m *DownloadMovieService) DownloadImg(URL string) (string, error) {

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
