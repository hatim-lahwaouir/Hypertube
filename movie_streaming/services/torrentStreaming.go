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
	PieceLength int64
	Pieces      [][20]byte
	Bitfield    []byte

	Files []string
	Size  []int64

	HttpTrackers []string
	UdpTrackers  []bittorent.UdpTracker

	Downloaded int64
	Left       int64
    Peers      map[string]*bittorent.Peer
}

func (t *TorrentStreaming) GetState() *bittorent.CurrentState {

	return &bittorent.CurrentState{Downloaded: t.Downloaded, Left: t.Left, Port: t.Port, PeerId: t.PeerId, InfoHash: t.InfoHash}
}

func (t *TorrentStreaming) DownloadAPiece(cur uint32)  error {
    // iterate over all the peers
    // ask them for the block of this piece

    // we need something to parse the reponse and also to check if this block of piece already exists


    var (
     //   piece_length uint32
        blockSize uint32
        base uint32
        //n_block uint32
      //  downloaded uint32
    )
    //downloaded =  0
    //piece_length = uint32(t.PieceLength)
    blockSize = 16384
    base = 16384
    maxPacketsToSend := 5 

    
        //n_block = (pice_length + (blockSize - 1))  / blockSize


        //for downloaded < piece_length{
            for _, val := range(t.Peers){ 
                if val.UnChoke {
                        fmt.Println("good peer")
                }
                if val.UnChoke == false || val.IsGood() == false ||   val.PacketSent >= maxPacketsToSend {
                    continue
                }
                if err := val.Request(cur, base,blockSize); err != nil {
                        fmt.Println("erroring sending piece request , ", err.Error())

                }

            }
            base += blockSize
       // }

    return nil
}



// goal of this is to download the movie
func (t *TorrentStreaming) MonitorPeers(wg *sync.WaitGroup){

    var (
        PeerWg sync.WaitGroup
    )
      
    defer wg.Done()
    t.GetUdpPeers()

    fmt.Println("--- start peer gorotoine--")
    for _, val := range(t.Peers){
        val.SetInfo(t.InfoHash, t.PeerId)
        val.InitBitField(int((((t.Length + (t.PieceLength - 1)) / t.PieceLength) + 7) / 8))
        PeerWg.Add(1)
        go val.PearGoRotine(&PeerWg)
    }
    




    PeerWg.Wait()
}







func NewTorrentStreaming(p string, t *bittorent.TorrentFile) *TorrentStreaming {

	var (
		files        []string
		size         []int64
		pieces       [][20]byte
		httpTrackers []string
		udpTrackers  []bittorent.UdpTracker
		peerId       [20]byte
		bitfield     []byte
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

    n_piece := (t.CalculateLength() + (t.Info.PieceLength  - 1)  ) /  t.Info.PieceLength
    bitfield = make([]byte, (n_piece + 7) / 8) 

    
	portUint64, _ := strconv.ParseUint(p, 10, 16)
	return &TorrentStreaming{PeerId: peerId, InfoHash: [20]byte(t.InfoHash), Port: uint16(portUint64), Length: t.CalculateLength(),
		Size: size, Files: files, UdpTrackers: udpTrackers, HttpTrackers: httpTrackers, Left: t.CalculateLength(), Downloaded: 0, PieceLength : t.Info.PieceLength, Bitfield: bitfield, Peers : make(map[string]*bittorent.Peer)}
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
    //t.Peers = peers
    for _ , p := range peers {
        t.Peers[p.Id()] = &p
    }
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



func (t *TorrentStreaming) GetGoodPeers(res chan  *bittorent.Peer, wg *sync.WaitGroup){
    defer wg.Done()
    var (
        newPeers  []*bittorent.Peer
    )

    for p := range(res) {
        newPeers = append(newPeers,p)
    }
   // t.Peers = newPeers
    for _ , p := range newPeers{
        t.Peers[p.Id()] = p
    }
}



