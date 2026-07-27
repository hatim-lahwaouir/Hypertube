package services

import (
	"fmt"
	"io"
	"os"
	"sync"

	bittorent "github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
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

func (ms *MovieStreamingService) ParseTorrent(r io.Reader) error {
	t, err := bittorent.NewTorrent(r)
	if err != nil {
		return err
	}
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
