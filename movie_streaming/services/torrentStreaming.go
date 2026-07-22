package services

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	bittorent "github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
)

type TorrentStreaming struct {
	PeerId      [20]byte
	InfoHash    [20]byte
	Port        uint16
	Length      int64
	PieceLength int64
	Pieces      [][20]byte
	Bitfield    []byte

	Files []bittorent.File

	HttpTrackers []string
	UdpTrackers  []bittorent.UdpTracker

	Downloaded int64
	Left       int64
    Peers      map[string]*bittorent.Peer
	NPiece     uint32
	MoviePath  string

	PieceWorkRecvChan  chan *bittorent.PieceWork
	PieceWorkResChan chan *bittorent.PieceWork
}


func NewTorrentStreaming(p string, t *bittorent.TorrentFile) *TorrentStreaming {

	var (
		files        []bittorent.File
		pieces       [][20]byte
		httpTrackers []string
		udpTrackers  []bittorent.UdpTracker
		peerId       [20]byte
		bitfield     []byte
	)

	// incase there is no tracker
	trackers := []string{
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://tracker.torrent.eu.org:451/announce",
	"udp://tracker.dler.org:6969/announce",
	"udp://open.stealth.si:80/announce",
	"udp://open.demonii.com:1337/announce",
	"https://tracker.moeblog.cn:443/announce",
	"udp://open.dstud.io:6969/announce",
	"udp://tracker.srv00.com:6969/announce",
	"https://tracker.zhuqiy.com:443/announce",
	"https://tracker.pmman.tech:443/announce",
	}

	for _, val := range(trackers){
		t.AnnounceList = append(t.AnnounceList, []string{val})
	}
	for _, val := range t.Info.Files {
		files = append(files, *bittorent.Newfile(filepath.Join(val.Path...), uint32(val.Length)))
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
		 Files: files, UdpTrackers: udpTrackers, HttpTrackers: httpTrackers, Left: t.CalculateLength(), Downloaded: 0, PieceLength : t.Info.PieceLength, 
		Bitfield: bitfield, Peers : make(map[string]*bittorent.Peer),
		NPiece: uint32(n_piece), Pieces: pieces, MoviePath: os.Getenv("MOVIES_PATH"), PieceWorkRecvChan  : make(chan *bittorent.PieceWork, 10),
	PieceWorkResChan  : make(chan *bittorent.PieceWork, 10)}
}

func (t *TorrentStreaming) GetState() *bittorent.CurrentState {

	return &bittorent.CurrentState{Downloaded: t.Downloaded, Left: t.Left, Port: t.Port, PeerId: t.PeerId, InfoHash: t.InfoHash}
}


func (t *TorrentStreaming) DeleteAbandonedPeers(){
		for _, p := range(t.Peers){
			if !p.IsGood(){
				p.Clear()
				delete(t.Peers, p.Id())
			}
		}
}


func (t *TorrentStreaming) WritePicesIntoTheDisk(wg *sync.WaitGroup, pieces chan bittorent.PieceWork){
	defer wg.Done()
	curFile := 0

	var (
		data []byte
	)
	MovieDirectory := filepath.Join(t.MoviePath, fmt.Sprintf("%s-%d-movie", uuid.NewString(), time.Now().Unix()))
	err := os.MkdirAll(MovieDirectory, 0755)
	if err != nil {
		fmt.Println("err", err)
	}
	
	for piece := range(pieces){

		if curFile >= len(t.Files){
			fmt.Println("the whole torrent is downloaded ")
			return
		}
		file := t.Files[curFile]

		if !file.IsCreated(){
			err := file.Create(MovieDirectory)
			fmt.Println(">>>>creating file err ", err, MovieDirectory)
		}

		data = append(data, piece.Buffer...)


		overlflow , err := file.WriteData(int(piece.Index),data)
		if err == nil {
			continue
		}
		if len(overlflow) != 0{
			data = append(data, piece.Buffer...)
		}
		if file.Done(){
			curFile += 1
		}
	}

}


func (t *TorrentStreaming) PrintPeers(){

	fmt.Println("-- peers -- ")
	for _, p := range(t.Peers){
		fmt.Println(p.Id())
	}
}

func (t *TorrentStreaming) StartPeers(wg *sync.WaitGroup){
    
	for _, val := range(t.Peers){
		if val.Started {
			continue
		}
		fmt.Println("starting peers", val.Id())
		val.Started = true
		val.SetInfo(t.InfoHash, t.PeerId)
        val.InitBitField(int((((t.Length + (t.PieceLength - 1)) / t.PieceLength) + 7) / 8))
        val.SetChannel(t.PieceWorkRecvChan , t.PieceWorkResChan)
		wg.Add(1)
        go val.PeerGoRotine(wg)
    }
}
// MonitorPeers goal of this is to download the movie
func (t *TorrentStreaming) MonitorPeers(wg *sync.WaitGroup){

    var (
		// piece_send int
        PeerWg sync.WaitGroup
		piecesToWrite chan bittorent.PieceWork

    )

      
	defer wg.Done()
	piecesToWrite = make(chan  bittorent.PieceWork, 20)

    t.GetUdpPeers()

	go t.WritePicesIntoTheDisk(&PeerWg, piecesToWrite)

	t.StartPeers(&PeerWg)

	// clear connections 
	// t.DeleteAbandonedPeers()

	// curPiece := uint32(0)

	windowPieces := 5
	for curPiece := int64(0); uint32(curPiece) < (t.NPiece); curPiece += int64(windowPieces){

		
		
		// if (curPiece + 1) % 3 == 0{
		// 	// get New peers
		// 	fmt.Print("get new peers")
		// 	t.GetUdpPeers()
		// 	t.StartPeers(&PeerWg)
		// 	t.PrintPeers()
		// }
		if windowPieces + int(curPiece) > int(t.NPiece){
			windowPieces = int(t.NPiece) - int(curPiece)
		}

		for i := curPiece; i < curPiece + int64(windowPieces); i += 1{
			fmt.Println("sending peice", i )
			t.PieceWorkRecvChan <- bittorent.NewPieceWork(uint32(i), uint32(t.PieceLength), uint32(t.Length), t.Pieces[curPiece])
		}
		// for _, p := range(t.Peers){
		// 		if p.IsGood(){
		// 			select {
		// 			case p.PieceWorkRecvChan <- piece:
		// 			default:
		// 			}
		// 		}
		// }
		n := 0
		for  n < windowPieces {
			select{
			case PieceRes := <- t.PieceWorkResChan:
				fmt.Println(PieceRes.Index, "reciving piece")
				 status := PieceRes.ValidateEntigrity()
				if !status{
					t.PieceWorkRecvChan <- PieceRes
				}else{
					piecesToWrite <- *PieceRes
					n += 1
				}
			default:
				// fmt.Println("nothing was recived")
				time.Sleep(time.Millisecond * 200)
			}

		}

	// for !piece.Done() {
	// 	// piece_send =0 
	// 	piece.PrintState()
	// 	t.DeleteAbandonedPeers()
	// }
	
	// if !piece.ValidateEntigrity(){
	// 		fmt.Println("piece ", curPiece, "failled entigrity checks")		
	// 		curPiece -= 1
	// 		continue
	// }
	// i have this piece 
			
	} 
    PeerWg.Wait()
}



func (t *TorrentStreaming) ParsePiece(p *bittorent.PieceWork, wg *sync.WaitGroup){


	p.Downloaded += p.Size
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


func (t *TorrentStreaming) GetHTTPPeers(){

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
    n_gorotines = 10


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



