package services

import (
	"crypto/rand"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "sync"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type TorrentStreaming struct {
	PeerId      [20]byte
	InfoHash    [20]byte
	Port        uint16
	Length      int64
	PieceLength int
	Pieces      [][20]byte

	Files []string
	Size  []int64

	HttpTrackers []string
	UdpTrackers  []types.UdpTracker

	Downloaded int64
	Left       int64
    Peers      []types.Peer
}

func (t *TorrentStreaming) GetState() *types.CurrentState {

	return &types.CurrentState{Downloaded: t.Downloaded, Left: t.Left, Port: t.Port, PeerId: t.PeerId, InfoHash: t.InfoHash}

}

func NewTorrentStreaming(p string, t *types.TorrentFile) *TorrentStreaming {

	var (
		files        []string
		size         []int64
		pieces       [][20]byte
		httpTrackers []string
		udpTrackers  []types.UdpTracker
		peerId       [20]byte
	)

	for _, val := range t.Info.Files {
		files = append(files, filepath.Join(val.Path...))
		size = append(size, int64(val.Length))
	}

	hashSize := 20
	piecesByte := []byte(t.Info.Pieces)
	numPieces := len(piecesByte) / hashSize

	for i := 0; i < numPieces; i += 1 {

		start := i * hashSize
		end := start + hashSize

		b := [20]byte(piecesByte[start:end])
		pieces = append(pieces, b)
	}

	if strings.HasPrefix(t.Announce, "http:") || strings.HasPrefix(t.Announce, "https:") {
		httpTrackers = append(httpTrackers, t.Announce)
	} else {

		udpTrackers = append(udpTrackers, types.NewUdpTracker(t.Announce))
	}

	for _, values := range t.AnnounceList {

		for _, val := range values {

			if strings.HasPrefix(val, "http:") || strings.HasPrefix(val, "https:") {
				httpTrackers = append(httpTrackers, val)
			} else {

				udpTrackers = append(udpTrackers, types.NewUdpTracker(val))
			}
		}
	}

	if _, err := rand.Read(peerId[:]); err != nil {
		peerId = [20]byte{72, 101, 108, 108, 111}
	}

	portUint64, _ := strconv.ParseUint(p, 10, 16)
	return &TorrentStreaming{PeerId: peerId, InfoHash: [20]byte(t.InfoHash), Port: uint16(portUint64), Length: t.CalculateLength(),
		Size: size, Files: files, UdpTrackers: udpTrackers, HttpTrackers: httpTrackers, Left: t.CalculateLength(), Downloaded: 0}
}

func (t *TorrentStreaming) GetPeers(){

    var (
        peers []types.Peer

    )

	// using http trackers
	cur := t.GetState()

	for _, u := range t.HttpTrackers {

		base, _ := url.Parse(u)

		params := url.Values{
			"info_hash":  []string{string(t.InfoHash[:])},
			"peer_id":    []string{string(t.PeerId[:])},
			"port":       []string{strconv.FormatUint(uint64(t.Port), 10)},
			"uploaded":   []string{"0"},
			"downloaded": []string{strconv.FormatInt(cur.Downloaded, 10)},
			"compact":    []string{"1"},
			"left":       []string{strconv.FormatInt(cur.Left, 10)},
		}
		base.RawQuery = params.Encode()
	}

	for _, u := range t.UdpTrackers {
       
		u.GetConnectionId()
        p := u.GetPeers(cur)

        if p != nil {
		    peers = append(peers, p...)
        }
	}
    t.Peers = peers
}


func TryHandShake(h types.HandShake, recv chan types.Peer, res chan types.Peer,wg *sync.WaitGroup) {

    defer wg.Done()
    for p := range(recv) {
        p.Connect()
        if p.IsGood == false{
            continue
        }
        p.PeerHandShake(h)
        if p.IsGood == false{
            continue
        }
        res <- p
    }
}


func (t *TorrentStreaming) HandShake(){
    //PeerId      [20]byte
	//InfoHash    [20]byte
    var (
        //GoodPeers []types.Peer
        res    chan types.Peer
        recv    chan types.Peer
        wg sync.WaitGroup
        n_gorotines int

    )

    n_gorotines = 100


    h := types.NewHandShake(t.PeerId, t.InfoHash)
    res = make(chan types.Peer, 100)
    recv = make(chan types.Peer, 100)

    for i := 0; i <n_gorotines; i += 1 {
        wg.Add(1)
        go TryHandShake(*h, recv, res, &wg) 
    }


    for _, p := range(t.Peers) {
        recv <- p
    }
    close(recv)
}
