package services

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

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

func (ms *MovieStreamingService) ParseTorrent(r io.Reader) (string, error) {
	t, err := bittorent.NewTorrent(r)
	if err != nil {
		return "", err
	}
	ts := NewTorrentStreaming("6881", t)

    ms.wg.Add(1)
    go  ts.MonitorPeers(&ms.wg)
    ms.AddStream(ts)

	
    return hex.EncodeToString(ts.InfoHash[:]) , nil
}

func (ms *MovieStreamingService) AddStream(stream *TorrentStreaming) {
    _ , ok := ms.Streams[stream.InfoHash]

    if !ok {
        ms.Streams[stream.InfoHash] = stream
    }
}



func (ms *MovieStreamingService) HasBitField(infoHash []byte, start uint64, end uint64) ([]byte, error) {
	// 1- check if server has the pieces requested
	// 2- get the pieces requested and stream them to the client 
    t , ok := ms.Streams[[20]byte(infoHash)]

	if !ok{
		return nil, nil 
	}
	if end >= uint64(t.MovieFile.Size){
		end = uint64(t.MovieFile.Size) - 1
	}


	for t.HasRange(int64(start), int64(end)) == false {
		time.Sleep(300 * time.Millisecond)
		fmt.Println("------------------still not good")
	}

	data, err := t.ReadChunk(int(start), int(end - start) + 1)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	return data, nil
}


func (ms *MovieStreamingService) FileSize(infoHash []byte) (int) {
	// 1- check if server has the pieces requested
	// 2- get the pieces requested and stream them to the client 
    t , ok := ms.Streams[[20]byte(infoHash)]

	if !ok{
		return -1
	}

	return int(t.MovieFile.Size)
}