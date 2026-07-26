package services

import (
	"crypto/rand"
	"fmt"
	"maps"
	mrand "math/rand"
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

	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
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
	MovieFile *bittorent.File

	HttpTrackers []string
	UdpTrackers  *bittorent.UdpTrackers

	Downloaded int64
	Left       int64
    Peers      map[string]*bittorent.Peer
	PeersMutex sync.RWMutex
	NPiece     uint32
	

	FailledPiece  chan *bittorent.PieceWork
	PieceWorkResChan chan *bittorent.PieceWork
}




func GetMovieFileOnly(files []*bittorent.File) *bittorent.File{

	var chosenFile *bittorent.File
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Size > files[j].Size
	})

	videoExts := map[string]bool{
        ".mp4": true, ".mkv": true, ".avi": true,
        ".mov": true, ".webm": true, ".m4v": true,
    }

	chosenFile = files[0]
	for i := range(files){
		ext := filepath.Ext(files[i].FilePath)
		if videoExts[ext]{
			chosenFile = files[i]
			break
		}
	}

	return chosenFile
}

func (t *TorrentStreaming) FirstPiece() int64 {
	return t.MovieFile.Offset / t.PieceLength
}

func (t *TorrentStreaming) LastPiece() int64 {
	return (t.MovieFile.Offset + t.MovieFile.Size - 1) / t.PieceLength
}


func NewTorrentStreaming(p string, t *bittorent.TorrentFile) *TorrentStreaming {

	var (
		files        []*bittorent.File
		pieces       [][20]byte
		httpTrackers []string
		udpTrackers  []*bittorent.UdpTracker
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

	GetMovieFileOnly(files)
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

    var n_piece int64 = (t.CalculateLength() + (t.Info.PieceLength  - 1)  ) /  t.Info.PieceLength
    bitfield = make([]byte, (n_piece + 7) / 8) 



	portUint64, _ := strconv.ParseUint(p, 10, 16)
	return &TorrentStreaming{PeerId: peerId, InfoHash: [20]byte(t.InfoHash), Port: uint16(portUint64), Length: t.CalculateLength(),
		 Files: files, UdpTrackers: bittorentProtocol.NewUdpTrackers(udpTrackers) , HttpTrackers: httpTrackers, Left: t.CalculateLength(), Downloaded: 0, PieceLength : t.Info.PieceLength, 
		Bitfield: bitfield, Peers : make(map[string]*bittorent.Peer),
		NPiece: uint32(n_piece), Pieces: pieces,
	PieceWorkResChan  : make(chan *bittorent.PieceWork, 10), MovieFile: GetMovieFileOnly(files)}
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
		start int64
		end int64
	)

	f := t.MovieFile
	if err := f.Create(); err != nil {
		fmt.Println("err creating movie file", err)
		os.Exit(1)
	}

	for piece := range(pieces){
		globalOffset := int64((piece.Index) * uint32(t.PieceLength))
		data = piece.Buffer
		if f.Done(){
				break
		}

		
		start = 0 
		if globalOffset < f.Offset {
			start = f.Offset - globalOffset
		}

		end = int64(len(data))
		if globalOffset + int64(piece.Size) > (f.Offset + f.Size){
			end = int64(len(data)) -  ((globalOffset + int64(piece.Size)) - (f.Offset + f.Size)) 
		}

		data = data[start:end]

		localOffset := (globalOffset + start ) - f.Offset
		if err := f.WriteData(localOffset, data);err != nil {
						fmt.Println("err writing at a file", err)
		}
	}
}

		// if f.Offset <= globalOffset && globalOffset < f.Offset + f.Size {
		// 		bytesToWrite := int64(len(data)) 
				
		// 		if globalOffset + int64(len(data)) > f.Offset + f.Size {
		// 			bytesToWrite =  f.Size - localOffset 
		// 		}
		// 		if !f.IsCreated(){
		// 			if err := f.Create(); err != nil {
		// 				fmt.Println("err creating file", err)
		// 			}
		// 		}
		// 		fmt.Println("writing data of piece", piece.Index)
	

		// 		data = data[bytesToWrite:]
		// 		globalOffset += bytesToWrite
		// }

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
		select {
		case t.Peers[p].BroadCastMsg <- msg:
		default:
			continue
		}

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
					t.PeersMutex.RLock()
					Peers := slices.Collect(maps.Values(t.Peers))
					t.PeersMutex.RUnlock()
					sort.Slice(Peers, func(i, j int) bool {
							return Peers[i].GetBytesReceived() > Peers[j].GetBytesReceived()
						})
					for i := range(Peers){
						if i < 4 && Peers[i].IsInterested() {
							fmt.Println("peer is interested")
							Peers[i].UnChokePeer(true)
						}else if Peers[i] != optimisticPeer{
							Peers[i].UnChokePeer(false)
						}
						Peers[i].ResetBytesReceived()
					}

				case <- optimisticTicker.C:
						var chokedPeers []*bittorent.Peer
						t.PeersMutex.RLock() 
						for i := range t.Peers {
							if t.Peers[i].IsGood() && t.Peers[i].ChokeUploadStatus(){
								chokedPeers = append(chokedPeers, t.Peers[i])
							}
						}
						t.PeersMutex.RUnlock()

						if len(chokedPeers) > 0 {
							randomIndex := mrand.Intn(len(chokedPeers))
							chosen := chokedPeers[randomIndex]
							
							chosen.UnChokePeer(true)
							optimisticPeer = chosen 
						} else {
							optimisticPeer = nil
						}
					t.GetUDPPeers()
					t.StartPeers(wg)
					t.DeleteAbandonedPeers()
			}
	}

}


func (t *TorrentStreaming) GetUDPPeers(){
	newPeers := t.UdpTrackers.GetPeers(t.GetState())


	t.PeersMutex.Lock()
	for i := range newPeers {
	
		if _, exists := t.Peers[newPeers[i].ID()]; exists {
			continue
		}
        t.Peers[newPeers[i].ID()] = newPeers[i]
    }
	t.PeersMutex.Unlock()
}


func (t *TorrentStreaming) CalculateTheWindow() int64 {
	targetWindowBytes := int64(15 * 1024 * 1024) 
    
    windowPieces := (targetWindowBytes + t.PieceLength - 1) / t.PieceLength
    if windowPieces < 1 {
        windowPieces = 1
    }

	return windowPieces
}



// send piece if peer has it 
// if he failled to install it 
// need to send it back 


func (t *TorrentStreaming) SendPiece(piece *bittorent.PieceWork) {

	t.PeersMutex.RLock()
	defer t.PeersMutex.RUnlock()

	for i := range(t.Peers){
		if !t.Peers[i].IsGood(){
			continue
		} 
		
		if t.Peers[i].HasPiece(piece.Index){
			select{
			case t.Peers[i].PieceWorkRecvChan <- piece:
				return 
			default:
				continue
			}
		}
	}

	t.FailledPiece <- piece
}

func (t *TorrentStreaming) MonitorPeers(wg *sync.WaitGroup){

    var (
		// piece_send int
        PeerWg sync.WaitGroup
		piecesToWrite chan *bittorent.PieceWork
		piecesDownloded int64
		inflight int64

    )

      
	//
	t.FailledPiece = make(chan *bittorent.PieceWork , t.NPiece / 3)
	defer wg.Done()
	piecesToWrite = make(chan  *bittorent.PieceWork, 20)

    
	PeerWg.Add(1)
	go t.WritePicesIntoTheDisk(&PeerWg, piecesToWrite)
	PeerWg.Add(1)
 	go t.TitForTatIshAlgho(&PeerWg)

	t.GetUDPPeers()
	t.StartPeers(&PeerWg)

	piecesDownloded = 0
	inflight = 0
	nextPieceToRequest := t.FirstPiece()

	start := time.Now()
	windowPieces := t.CalculateTheWindow()
	for piecesDownloded < t.LastPiece() - t.FirstPiece(){		
		for inflight < windowPieces && nextPieceToRequest < t.LastPiece(){
			t.SendPiece(bittorent.NewPieceWork(uint32(nextPieceToRequest), uint32(t.PieceLength), uint32(t.Length), t.Pieces[nextPieceToRequest]))
			inflight++
			nextPieceToRequest++
		}


			select{
			case PieceRes := <- t.PieceWorkResChan:

				if !PieceRes.ValidateEntigrity(){
					t.FailledPiece <- PieceRes
				}else{
					piecesToWrite <- PieceRes
					//
					t.BroadCastHaveMsg(PieceRes.Index)
					t.Downloaded += int64(PieceRes.Size)
					t.SetPiece(PieceRes.Index)
					inflight--
					piecesDownloded++
				}
			case failledPeice := <- t.FailledPiece:
				t.SendPiece(failledPeice)
			default:
			
			}

			time.Sleep(100 * time.Millisecond)
			if nextPieceToRequest % 10 == 0 {
				fmt.Printf("[%.2f | 100%%]\n", 100 * float64(t.Downloaded) / float64(t.MovieFile.Size))
				fmt.Println(piecesDownloded, "were downloded in", time.Until(start).Abs())

				
			}
		}
	
	
		fmt.Println("we are done !")
		close(piecesToWrite)
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
	fmt.Println("cur peers", len(t.Peers))
	for _, val := range(t.Peers){
		if val.Started {
			continue
		}
		val.Started = true
		val.SetInfo(t.InfoHash, t.PeerId)
        val.InitBitField(int((((t.Length + (t.PieceLength - 1)) / t.PieceLength) + 7) / 8))
        val.SetChannel(t.FailledPiece , t.PieceWorkResChan)
		val.InitBroadcast()
		if t.Downloaded > 0 { 
			val.SetUpServerBitField(t.Bitfield)
		}
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
