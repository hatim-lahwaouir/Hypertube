package services

import (
	"path/filepath"
    "fmt"
	"compress/gzip"
	"encoding/hex"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
	bittorent "github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
	"os"
    "sync"
)

type MovieStreamingService struct {
	TorrentPath string
	Streams     map[[20]byte]*TorrentStreaming 
    wg sync.WaitGroup
}

var cacheMovieStreaming *MovieStreamingService

func NewMovieStreamingService() *MovieStreamingService {
	if cacheMovieStreaming == nil {
		cacheMovieStreaming = &MovieStreamingService{
			TorrentPath: os.Getenv("TORRENT_PATH"),
            Streams: make(map[[20]byte]*TorrentStreaming),
            
		}
	}
	return cacheMovieStreaming
}

func (ms *MovieStreamingService) ParseTorrent(torrent *models.Torrent) error {
	filePath := filepath.Join(ms.TorrentPath, torrent.Path)

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	t, err := bittorent.NewTorrent(gzipReader)
	if err != nil {
		return err
	}

	decodedByteArray, err := hex.DecodeString(torrent.Hash)
	if err != nil {
		return err
	}
	t.InfoHash = decodedByteArray
	ts := NewTorrentStreaming("6881", t)

    ms.wg.Add(1)
    go  ts.MonitorPeers(&ms.wg)
    ms.AddStream(ts)
    fmt.Println(ms.Streams)
    return nil
}

func (ms *MovieStreamingService) AddStream(stream *TorrentStreaming) {
    _ , ok := ms.Streams[stream.InfoHash]

    if !ok {
        ms.Streams[stream.InfoHash] = stream
    }
}
