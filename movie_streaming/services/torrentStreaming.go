package services

import (
    "fmt"
	"crypto/rand"
	bittorent "github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
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
	UdpTrackers  []bittorent.UdpTracker

	Downloaded int64
	Left       int64
    Peers      []bittorent.Peer
}

func (t *TorrentStreaming) GetState() *bittorent.CurrentState {

	return &bittorent.CurrentState{Downloaded: t.Downloaded, Left: t.Left, Port: t.Port, PeerId: t.PeerId, InfoHash: t.InfoHash}

}

func NewTorrentStreaming(p string, t *bittorent.TorrentFile) *TorrentStreaming {

	var (
		files        []string
		size         []int64
		pieces       [][20]byte
		httpTrackers []string
		udpTrackers  []bittorent.UdpTracker
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

		udpTrackers = append(udpTrackers, bittorent.NewUdpTracker(t.Announce))
	}

	for _, values := range t.AnnounceList {

		for _, val := range values {

			if strings.HasPrefix(val, "http:") || strings.HasPrefix(val, "https:") {
				httpTrackers = append(httpTrackers, val)
			} else {

				udpTrackers = append(udpTrackers, bittorent.NewUdpTracker(val))
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



func GetUdpPeersWorker(recv chan bittorent.UdpTracker, res chan []bittorent.Peer,  cur *bittorent.CurrentState ,wg *sync.WaitGroup) {
   

   defer wg.Done()

   for  tr := range(recv) {
        tr.GetConnectionId()
        p := tr.GetPeers(cur)
        res <- p
   }
}



func (t *TorrentStreaming) StorePeersFromUdpTracker(res chan []bittorent.Peer, wg *sync.WaitGroup){
    var (
        peers []bittorent.Peer
    )
    defer wg.Done()
    

    for p := range(res) {
        peers = append(peers, p...)
    }
    t.Peers = peers
}


func (t *TorrentStreaming) GetHttpPeers(){

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
    
}



func (t *TorrentStreaming) GetUdpPeers(){

    var (

        waitGetUdpPeers sync.WaitGroup
        waitWorkersUdpPeers sync.WaitGroup
        peerRes chan []bittorent.Peer
        recv chan bittorent.UdpTracker
        n_gorotines int
    )
    n_gorotines = 20


    recv = make(chan bittorent.UdpTracker, 50)
    peerRes = make(chan []bittorent.Peer, 100)

	// using http trackers
	cur := t.GetState()

    // udp
    for i := 0; i < n_gorotines; i += 1 {
        waitWorkersUdpPeers.Add(1)
        go GetUdpPeersWorker(recv, peerRes,cur , &waitWorkersUdpPeers)
    }
    waitGetUdpPeers.Add(1)
    go t.StorePeersFromUdpTracker(peerRes, &waitGetUdpPeers)
	for _, u := range t.UdpTrackers {
        recv <- u
	}
    close(recv)
    waitWorkersUdpPeers.Wait()
    close(peerRes)

    waitGetUdpPeers.Wait()
}


func TryHandShake(h bittorent.HandShake, recv chan  bittorent.Peer, res chan bittorent.Peer,wg *sync.WaitGroup) {

    defer wg.Done()
    for p := range(recv) {
        p.Connect()
        if p.IsGood == false{
            continue
        }
        resp := p.PeerHandShake(h)
        if p.IsGood == false{
            continue
        }
        if p.ValidHandShake(h, resp) == false {
            continue
        }

        fmt.Println("-- handshake --")
        fmt.Println("new peer handshake good ", p)
        res <- p
    }
}


func (t *TorrentStreaming) GetGoodPeers(res chan  bittorent.Peer, wg *sync.WaitGroup){
    defer wg.Done()
    var (
        newPeers  []bittorent.Peer
    )

    for p := range(res) {
        newPeers = append(newPeers,p)
    }
    t.Peers = newPeers
}

func (t *TorrentStreaming) HandShake(){
    var (
        //GoodPeers []bitt.Peer
        res    chan bittorent.Peer
        recv    chan bittorent.Peer
        waitWorkers sync.WaitGroup
        waitGoodPeers sync.WaitGroup
        n_gorotines int

    )
    n_gorotines = 50

    h := bittorent.NewHandShake(t.PeerId, t.InfoHash)
    res = make(chan bittorent.Peer, 300)
    recv = make(chan bittorent.Peer, 300)

    for i := 0; i <n_gorotines; i += 1 {
        waitWorkers.Add(1)
        go TryHandShake(*h, recv, res, &waitWorkers) 
    }
    waitGoodPeers.Add(1)
    go t.GetGoodPeers(res, &waitGoodPeers)
    

    for _, p := range(t.Peers) {
        recv <- p
    }


    close(recv)
    waitWorkers.Wait()
    close(res)
    waitGoodPeers.Wait()
    fmt.Println(len(t.Peers), "we got n peers ")
}
