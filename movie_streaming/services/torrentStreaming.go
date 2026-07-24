package services

import (
"math/rand"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
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

	Files []*bittorent.File

	HttpTrackers []string
	UdpTrackers  []bittorent.UdpTracker

	Downloaded int64
	Left       int64
    Peers      map[string]*bittorent.Peer
	PeersMutex sync.RWMutex
	NPiece     uint32

	PieceWorkRecvChan  chan *bittorent.PieceWork
	PieceWorkResChan chan *bittorent.PieceWork
}


func NewTorrentStreaming(p string, t *bittorent.TorrentFile) *TorrentStreaming {

	var (
		files        []*bittorent.File
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
	offset := int64(0)
	moviePath := filepath.Join(os.Getenv("MOVIES_PATH"), fmt.Sprintf("%s-%d-movie", uuid.NewString(), time.Now().Unix()))
	for _, val := range t.Info.Files {
		files = append(files, bittorent.Newfile(moviePath, filepath.Join(val.Path...), uint32(val.Length),t.Info.PieceLength, offset))
		offset += val.Length
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
		NPiece: uint32(n_piece), Pieces: pieces, PieceWorkRecvChan  : make(chan *bittorent.PieceWork, 10),
	PieceWorkResChan  : make(chan *bittorent.PieceWork, 10)}
}

func (t *TorrentStreaming) GetState() *bittorent.CurrentState {

	return &bittorent.CurrentState{Downloaded: t.Downloaded, Left: t.Left, Port: t.Port, PeerId: t.PeerId, InfoHash: t.InfoHash}
}


func (t *TorrentStreaming) DeleteAbandonedPeers(){

		t.PeersMutex.Lock()
		defer t.PeersMutex.Unlock()
		for _, p := range(t.Peers){
			if !p.IsGood(){
				p.Clear()
				delete(t.Peers, p.ID())
			}
		}
}


func (t *TorrentStreaming) WritePicesIntoTheDisk(wg *sync.WaitGroup, pieces chan *bittorent.PieceWork){
	defer wg.Done()

	var (
		data []byte
	)

	for piece := range(pieces){
		globalOffset := int64((piece.Index) * piece.Size)
		data = piece.Buffer
		for i := range(t.Files){
			// here we are at the file that the piece belongs to 
			f := t.Files[i]
			if len(data) == 0{
				break
			}
			if f.Offset <= globalOffset && globalOffset < f.Offset + f.Size {
				bytesToWrite := int64(len(data)) 
				localOffset := globalOffset - f.Offset
				if globalOffset + int64(len(data)) > f.Offset + f.Size {
					bytesToWrite =  f.Size - localOffset 
				}
				if !f.IsCreated(){
					if err := f.Create(); err != nil {
						fmt.Println("err creating file", err)
					}
				}
				fmt.Println("writing data of piece", piece.Index)
				if err := f.WriteData(localOffset, data[:bytesToWrite]);err != nil {
						fmt.Println("err writing at a file", err)
				}

				data = data[bytesToWrite:]
				globalOffset += bytesToWrite
			}
		}
	}

}


func (t *TorrentStreaming) PrintPeers(){

	fmt.Println("-- peers -- ")
	for _, p := range(t.Peers){
		fmt.Println(p.ID())
	}
}


// MonitorPeers goal of this is to download the movie


func (t *TorrentStreaming) BroadCastHaveMsg(index uint32){
	
	msg := bittorent.HaveMsg(index)
	t.PeersMutex.RLock()
	defer t.PeersMutex.RUnlock()
	for p := range(t.Peers){
		if !t.Peers[p].IsGood() {
			continue
		}
		t.Peers[p].BroadCastMsg <- msg
	}
}


func (t *TorrentStreaming) BroadCastBitField(){
	

	msg := bittorent.Msg{ID: bittorent.MsgBitfield, Payload: t.Bitfield}
	t.PeersMutex.Lock()
	defer t.PeersMutex.Unlock()
	for p := range(t.Peers){
		if !t.Peers[p].IsGood() && !t.Peers[p].HasBitField {
			continue
		}
		t.Peers[p].BroadCastMsg <- msg.Serialize()
		t.Peers[p].HasBitField = true
	}
}

func (t *TorrentStreaming) SetPiece(index uint32) {
	byteIndex := index / 8
	offset := index % 8
	t.Bitfield[byteIndex] |= 1 << (7 - offset)
}





func (t *TorrentStreaming)  TitForTatIshAlgho(wg *sync.WaitGroup){

	evalTicker := time.NewTicker(10 * time.Second)
    optimisticTicker := time.NewTicker(30 * time.Second)

	defer evalTicker.Stop()
	defer optimisticTicker.Stop()
	
	
	// here is the map t.Peers
	
	var optimisticPeer *bittorent.Peer = nil 
	for ;; {
		select {
				case <- evalTicker.C:
					// get the top 4 peers
					t.PeersMutex.RLock() // <--- READ LOCK
					Peers := slices.Collect(maps.Values(t.Peers))
					t.PeersMutex.RUnlock() // <--- READ LOCK
					sort.Slice(Peers, func(i, j int) bool {
							return Peers[i].GetBytesReceived() > Peers[j].GetBytesReceived()
						})
					for i,p := range(Peers){
						if i < 4 {
							p.UnChokePeer(true)
						}else if p != optimisticPeer{
							p.UnChokePeer(false)
						}
					}

				case <- optimisticTicker.C:
						var chokedPeers []*bittorent.Peer
						t.PeersMutex.RLock() 
						for _, p := range t.Peers {
							if p.IsGood() && p.ChokeUploadStatus(){
								chokedPeers = append(chokedPeers, p)
							}
						}
						t.PeersMutex.RUnlock()

						if len(chokedPeers) > 0 {
							randomIndex := rand.Intn(len(chokedPeers))
							chosen := chokedPeers[randomIndex]
							
							chosen.UnChokePeer(true)
							optimisticPeer = chosen 
						} else {
							optimisticPeer = nil
						}
			}
	}

}



func (t *TorrentStreaming) MonitorPeers(wg *sync.WaitGroup){

    var (
		// piece_send int
        PeerWg sync.WaitGroup
		piecesToWrite chan *bittorent.PieceWork

    )

      
	defer wg.Done()
	piecesToWrite = make(chan  *bittorent.PieceWork, 20)

    
	wg.Add(1)
	go t.WritePicesIntoTheDisk(&PeerWg, piecesToWrite)


	// clear connections 
	// t.DeleteAbandonedPeers()

	// curPiece := uint32(0)

	windowPieces := 10
	for curPiece := int64(0); uint32(curPiece) < (t.NPiece); curPiece += int64(windowPieces){
		
		

		t.GetUdpPeers()
		t.StartPeers(&PeerWg)

		t.BroadCastBitField()
		if windowPieces + int(curPiece) > int(t.NPiece){
			windowPieces = int(t.NPiece) - int(curPiece)
		}

		for i := curPiece; i < curPiece + int64(windowPieces); i += 1{

			t.PieceWorkRecvChan <- bittorent.NewPieceWork(uint32(i), uint32(t.PieceLength), uint32(t.Length), t.Pieces[i])
		}
		n := 0
		for  n < windowPieces {
			select{
			case PieceRes := <- t.PieceWorkResChan:
				fmt.Println(PieceRes.Index, "reciving piece")

				if !PieceRes.ValidateEntigrity(){
					fmt.Println("entigirity failled")
					t.PieceWorkRecvChan <- PieceRes
				}else{
					fmt.Println("entigirity succed")
					piecesToWrite <- PieceRes
					t.BroadCastHaveMsg(PieceRes.Index)
					n += 1
				}
			default:
				// fmt.Println("nothing was recived")
				t.DeleteAbandonedPeers()
				
				time.Sleep(time.Microsecond * 500)
			}

		}
		time.Sleep(1 * time.Second)
		fmt.Println("piece that are good", n)
			
	} 
    PeerWg.Wait()
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


// --------- peers functions -------------- 

func (t *TorrentStreaming) StartPeers(wg *sync.WaitGroup){
    
	t.PeersMutex.RLock()
	for _, val := range(t.Peers){
		if val.Started {
			continue
		}
		val.Started = true
		val.SetInfo(t.InfoHash, t.PeerId)
        val.InitBitField(int((((t.Length + (t.PieceLength - 1)) / t.PieceLength) + 7) / 8))
        val.SetChannel(t.PieceWorkRecvChan , t.PieceWorkResChan)
		val.InitBroadcast()
		var fileUploads []*bittorent.FileUploads
		
		for _, f := range(t.Files) {
			fileUploads = append(fileUploads, f.NewFileUploads())
		}
		val.ChokeUpload = true 
		val.SetUpFileUploads(fileUploads)
		wg.Add(1)
        go val.PeerGoRotine(wg)
    }
	t.PeersMutex.RUnlock()

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
	t.PeersMutex.Lock()
    for _ , p := range peers {
        t.Peers[p.ID()] = &p
    }
	t.PeersMutex.Unlock()
}



