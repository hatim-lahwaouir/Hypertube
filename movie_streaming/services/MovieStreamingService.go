package services

import (
    "path/filepath"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "compress/gzip"
    "fmt"
    "os"
)


type MovieStreamingService struct {
    TorrentPath string
}




var cacheMovieStreaming *MovieStreamingService


func NewMovieStreamingService() *MovieStreamingService {
    if  cacheMovieStreaming == nil {
        cacheMovieStreaming = &MovieStreamingService{
            TorrentPath: os.Getenv("TORRENT_PATH"),
            }   
    }
    return cacheMovieStreaming
}




func (ms *MovieStreamingService) ParseTorrent(fileName string) error {
    filePath := filepath.Join(ms.TorrentPath, fileName)

    f, err := os.OpenFile(filePath, os.O_RDONLY,  0644)
    if err != nil {
        return err
    }
	defer f.Close()

    gzipReader, err := gzip.NewReader(f)
	if err != nil {
        return err
	}
	defer gzipReader.Close()

    t, err :=  types.NewTorrent(gzipReader)
	if err != nil {
        return err
	}

	fmt.Println(">>>> torrent", t)
    

    return nil
} 
