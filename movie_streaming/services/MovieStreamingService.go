package services

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"github.com/google/uuid"

	bittorent "github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
)

type MovieStreamingService struct {
	TorrentPath string
	Streams     map[string]*TorrentStreaming 
    wg sync.WaitGroup
}

var cacheMovieStreaming *MovieStreamingService

func NewMovieStreamingService() *MovieStreamingService {
	if cacheMovieStreaming == nil {
		cacheMovieStreaming = &MovieStreamingService{
			TorrentPath: os.Getenv("TORRENT_PATH"),
            Streams: make(map[string]*TorrentStreaming),
            
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
    

	
    return ms.AddStream(ts) , nil
}

func (ms *MovieStreamingService) AddStream(stream *TorrentStreaming) string {
    
	
	key := uuid.NewString()
	_ , ok := ms.Streams[key]

    if !ok {
        ms.Streams[key] = stream
    }

	return key
}



func (ms *MovieStreamingService) HasBitField(id string, start uint64, end uint64) ([]byte, error) {
	// 1- check if server has the pieces requested
	// 2- get the pieces requested and stream them to the client 
    t , ok := ms.Streams[id]

	if !ok{
		return nil, nil 
	}
	if end >= uint64(t.MovieFile.Size){
		end = uint64(t.MovieFile.Size) - 1
	}


	t.ChangePriority(int64(start))

	for t.HasRange(int64(start), int64(end)) == false {
		time.Sleep(1 * time.Second)
		fmt.Println("------------------still not good")
	}

	data, err := t.ReadChunk(int(start), int(end - start) + 1)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	return data, nil
}


func (ms *MovieStreamingService) FileSize(id string ) (int) {
	// 1- check if server has the pieces requested
	// 2- get the pieces requested and stream them to the client 
    t , ok := ms.Streams[id]

	if !ok{
		return -1
	}

	return int(t.MovieFile.Size)
}