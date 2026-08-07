package bittorentProtocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
   "sync/atomic"
	// "fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)



type Peer struct {
	IP          net.IP
	Port        uint16
	Conn        net.Conn
	BitField    []byte
	valid       atomic.Bool
	UnChoke     atomic.Bool
	PacketSent  int
	ChokeUpload atomic.Bool

	InfoHash [20]byte

	PeerMutex    sync.Mutex
	ConnMutext   sync.Mutex
	MessagesSent atomic.Int32
	ClientID     [20]byte

	PieceWorkRecvChan chan *PieceWork
	FailledPiece chan *PieceWork
	PieceWorkResChan  chan *PieceWork
	Started           bool
	BroadCastMsg      chan []byte

	CurPiece     *PieceWork
	Interested   atomic.Bool
	BytesRecived atomic.Uint32
	PieceDone   chan bool
	HasBitField  bool
	ServerBitField []byte
	gone atomic.Bool

}






func (p *Peer) SetUpServerBitField(b []byte) {
	p.ServerBitField = b
}



func (p *Peer) RecivedBytes(n uint32) {
	p.BytesRecived.Add(n)
}

func (p *Peer) ResetBytesReceived() {
	p.BytesRecived.Store(0)
}

func (p *Peer) UnChokePeer(status bool) {
	if !p.IsGood(){
		return
	}

	p.ConnMutext.Lock()
	defer p.ConnMutext.Unlock()



	if p.ChokeUploadStatus() == !status {
		return
	}

	p.UpdateChokeUpload(!status)

	state := MsgChoke
	if status {
		state = MsgUnchoke
	}

	m := Msg{ID: state}
	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	if _, err := p.Conn.Write(m.Serialize()); err != nil {
		p.SetGood(false)
	}
}

func (p *Peer) ChokeUploadStatus() bool {
	return p.ChokeUpload.Load()
}

func (p *Peer) UpdateChokeUpload(status bool) {
	p.ChokeUpload.Store(status)
}

func (p *Peer) GetBytesReceived() uint32 {
	return p.BytesRecived.Load()
}

func (p *Peer) InterestedStatus(status bool) {
	p.Interested.Store(status)
}

func (p *Peer) IsInterested() bool {
	return p.Interested.Load()
}

func (p *Peer) SetChannel(FailledPiece chan *PieceWork, PieceWorkResChan chan *PieceWork) {
	p.FailledPiece = FailledPiece
	p.PieceWorkResChan = PieceWorkResChan
	p.PieceWorkRecvChan = make(chan *PieceWork, 5)
	p.PieceDone = make(chan bool, 1)
}

func (p *Peer) SentMessage() {
	p.MessagesSent.Add(1)
}

func (p *Peer) ReceivedMessage() {
	p.MessagesSent.Add(-1)
	if p.MessagesSent.Load() < 0 {
		p.MessagesSent.Store(0)
	}
}

func (p *Peer) CanSendMessage() bool {

	return p.MessagesSent.Load() < 32
}


func (p *Peer) SetGood(b bool) {
	p.valid.Store(b)
}

func (p *Peer) IsGood() bool {
	return p.valid.Load()
}

func (p *Peer) ID() string {
	return p.IP.String() + ":" + strconv.FormatUint(uint64(p.Port), 10)

}

func (p *Peer) SetInfo(infoHash [20]byte, clientID [20]byte) {
	p.InfoHash = infoHash
	p.ClientID = clientID
}

func (p *Peer) HasPiece(index uint32) bool {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	if p.BitField == nil {
		return false
	}
	byteIndex := index / 8
	offset := index % 8
	return p.BitField[byteIndex]>>(7-offset)&1 != 0
}

func (p *Peer) SetPiece(index uint32) {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	byteIndex := index / 8
	offset := index % 8
	if p != nil {
		p.BitField[byteIndex] |= 1 << (7 - offset)
	}
}

func NewPeers(resp []byte, n int) []*Peer {
	var peers []*Peer

	fmt.Println(len(resp), n)
	peersBinary := resp[20:n]
	peerSize := 6
	numPeers := len(peersBinary) / peerSize

	for i := 0; i < numPeers; i++ {
		offset := i * peerSize

		// Extract 4 bytes for IP and 2 bytes for Port
		ipBytes := peersBinary[offset : offset+4]
		portBytes := peersBinary[offset+4 : offset+6]

		// Convert port bytes to uint16 (Big Endian)
		port := binary.BigEndian.Uint16(portBytes)
		ip := net.IP(ipBytes)

		peers = append(peers, &Peer{IP: ip, Port: port, valid: atomic.Bool{}})
	}

	return peers
}

func (p *Peer) Connect() {
	conn, err := net.DialTimeout("tcp", p.IP.String()+":"+strconv.FormatUint(uint64(p.Port), 10), 10 *time.Second)
	if err != nil {
		p.SetGood(false)
		return
	}
	p.Conn = conn
}

func (p *Peer) PeerHandShake(h *HandShake) []byte {

	if !p.IsGood() {
		return nil
	}

	rawBytes := h.Serialize()

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := p.Conn.Write(rawBytes); err != nil {
		p.SetGood(false)
		return nil
	}
	resp := make([]byte, 68)

	p.Conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	_, err := p.Conn.Read(resp)
	if err != nil {
		p.SetGood(false)
		return nil
	}

	return resp
}

func (p *Peer) ValidHandShake(h *HandShake, peerResp []byte) bool {

	if !p.IsGood() {
		return false
	}
	rawBytes := h.Serialize()
	// comapre hash info and pstr
	if !bytes.Equal(peerResp[28:len(peerResp)-20], rawBytes[28:len(rawBytes)-20]) ||
		!bytes.Equal(rawBytes[0:20], peerResp[0:20]) {
		p.SetGood(false)
		return false
	}

	return true
}

func (p *Peer) GetMessage() (*Msg, error) {

	if !p.IsGood() {
		return nil, nil
	}

	p.Conn.SetDeadline(time.Now().Add(time.Second * 5))
	m, err := NewMessage(p.Conn)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return nil, nil
		}
		p.SetGood(false)
		return nil, err
	}
	return m, nil
}

func (p *Peer) Intersted() error {
	if !p.IsGood() {
		return nil
	}

	p.ConnMutext.Lock()
	defer p.ConnMutext.Unlock()
	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	m := Msg{ID: MsgInterested}
	data := m.Serialize()
	if _, err := p.Conn.Write(data); err != nil {
		p.SetGood(false)
		return err
	}

	return nil
}

func (p *Peer) Request(cur uint32, base uint32, length uint32) error {
	if !p.IsGood() || !p.UnChoke.Load() {
		return errors.New("invalid Peer")
	}
	buf := make([]byte, 17)

	binary.BigEndian.PutUint32(buf[0:4], 13)

	buf[4] = byte(MsgRequest)

	binary.BigEndian.PutUint32(buf[5:9], cur)

	binary.BigEndian.PutUint32(buf[9:13], base)

	binary.BigEndian.PutUint32(buf[13:17], length)

	p.ConnMutext.Lock()
	defer p.ConnMutext.Unlock()

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := p.Conn.Write(buf); err != nil {
		p.SetGood(false)
		return err
	}
	p.PacketSent++
	return nil

}

func (p *Peer) SendPiece(index uint32, begin uint32, piece []byte) error {
	if !p.IsGood() || p.ChokeUploadStatus() {
		return errors.New("invalid Peer")
	}
	payloadLen := (9 + len(piece))
	buf := make([]byte, 4+payloadLen)

	binary.BigEndian.PutUint32(buf[0:4], uint32(payloadLen))
	buf[4] = byte(MsgPiece)
	binary.BigEndian.PutUint32(buf[5:9], index)
	binary.BigEndian.PutUint32(buf[9:13], begin)
	copy(buf[13:], piece)

	p.ConnMutext.Lock()
	defer p.ConnMutext.Unlock()

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := p.Conn.Write(buf); err != nil {
		p.SetGood(false)
		return err
	}
	p.PacketSent++
	return nil

}

func (p *Peer) SetCurrentPiece(piece *PieceWork) {
	p.PeerMutex.Lock()
	p.CurPiece = piece
	p.PeerMutex.Unlock()
}

func (p *Peer) InitBroadcast() {
	p.BroadCastMsg = make(chan []byte, 10)
}

func (p *Peer) Broadcast(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range p.BroadCastMsg {
		p.ConnMutext.Lock()
		p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
		if _, err := p.Conn.Write(msg); err != nil {
			p.SetGood(false)
			p.ConnMutext.Unlock()
			return
		}
		p.ConnMutext.Unlock()
	}
}

func (p *Peer) SendBitField() error {


	if !p.IsGood() {
		return errors.New("client isn't good")
	}


	if p.ServerBitField == nil {
		return nil
	}

	msg := Msg{ID: MsgBitfield, Payload: p.ServerBitField}

	p.ConnMutext.Lock()
	defer p.ConnMutext.Unlock()

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := p.Conn.Write(msg.Serialize()); err != nil {
		p.SetGood(false)
		return err
	}

	return nil
}

func (p *Peer) PeerGoRotine(wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		other sync.WaitGroup
	)


	p.Connect()
	if !p.IsGood() {
		return
	}
	// first doing handshake with peer
	handShake := NewHandShake(p.ClientID, p.InfoHash)
	resp := p.PeerHandShake(handShake)

	if resp == nil {
		return
	}

	if !p.ValidHandShake(handShake, resp) {
		return
	}


	// here we will start a gorotine for reading peer messages
	if err := p.SendBitField(); err != nil {
		fmt.Println("errror setting bitfield", err)
	}

	if err := p.Intersted(); err != nil {
		fmt.Println("errror sending interest", err)
	}

	other.Add(2)
	go p.PeerMesgs(&other)
	go p.Broadcast(&other)

	for piece := range p.PieceWorkRecvChan {
		if !p.IsGood() {
			p.FailledPiece <- piece
			break
		}

		if !p.HasPiece(piece.Index) || p.PeerIsChoking() {
			p.FailledPiece <- piece
			continue
		}
		piecesChan := make(chan uint32, piece.NPiece+1)

		for begin := uint32(0); begin < piece.Size; begin += piece.BlockSize {
			piecesChan <- begin
		}

		p.ResetBytesReceived()
		p.SetCurrentPiece(piece)
		timeoutTicker := time.NewTicker(3 * time.Second)

		lastDownload := piece.GetDownloaded()
		stop := false
		for p.IsGood() && !stop {

			select {
			case begin := <-piecesChan:
				if piece.IsThisDone(begin) {
					continue
				}
				if p.PeerIsChoking(){
					stop = true
					continue
				}
				if !p.CanSendMessage(){
					piecesChan <- begin
					time.Sleep(30 * time.Millisecond)
					continue
				}

				length := piece.BlockSize
				if begin+length > piece.Size {
					length = piece.Size - begin
				}
				p.Request(piece.Index, begin, length)
				p.SentMessage()
			case <-timeoutTicker.C:
				if piece.Done(){
					continue
				}
				if lastDownload + 1000000 >  piece.GetDownloaded() {
					stop = true
				} else {
						lastDownload = piece.GetDownloaded()
				}
			case <- p.PieceDone:
				stop = true
			}
		}



		if !piece.Done(){
			p.FailledPiece <- piece

		} else{
			p.PieceWorkResChan <- piece
		}
		timeoutTicker.Stop()

		p.SetCurrentPiece(nil)
		close(piecesChan)
		
		if !p.IsGood() {
				break
		}
	}

	close(p.BroadCastMsg)
	other.Wait()

}

func (p *Peer) PeerChokingState(state bool) {

	p.UnChoke.Store(state)
}

func (p *Peer) PeerIsChoking() bool {
	return !p.UnChoke.Load()
}

func (p *Peer) PeerMesgs(wg *sync.WaitGroup) {
	defer wg.Done()
	if !p.IsGood() {
		return
	}
	for {
		if !p.IsGood() {
			break
		}
		m, err := p.GetMessage()
		if err != nil {
			p.SetGood(false)
			break
		}

		if m == nil {
			continue
		}
		switch m.ID {
		case MsgChoke:
			p.PeerChokingState(false)
		case MsgUnchoke:
			p.PeerChokingState(true)
		case MsgHave:
			if index, ok := m.ParseHave(); ok {
				p.SetPiece(index)
			}
		case MsgBitfield:
			if len(p.BitField) == len(m.Payload) {
				p.SetBitField(m.Payload) // 1MB 1GB -> 1000 + (7) / 8 Piece 
			}
		case MsgPiece:
			p.ReceivedMessage()
			index, begin, buf, ok := m.ParsePiece()
			// fmt.Println("Piece", index, "was received from ", p.ID())
			if !ok {
				continue
			}
			

			var piece *PieceWork = nil
			p.PeerMutex.Lock()
			if p.CurPiece != nil {
				piece = p.CurPiece
			}
			p.PeerMutex.Unlock()

			if piece == nil {
				continue
			}
			piece.SetPiece(index, buf, begin)
			p.RecivedBytes(uint32(len(buf)))
			if piece.Done(){
				select {
				case p.PieceDone <- true:
				default:
				}
			}
		case MsgInterested:
			fmt.Println("-------------------------send Interested ----------------------")
			p.InterestedStatus(true)
		case MsgRequest:
			fmt.Println("-------------------------send request ----------------------")
		case MsgExtention:
			fmt.Println("-------------------------Client support extension ----------------------")

		}

	}
	// here we will be waiting for peer messages
}

// Clear function to free all resources allocated
func (p *Peer) Clear() {

	if p.gone.Load(){
		return
	}
	p.gone.Store(true)
	p.valid.Store(false)


	p.ConnMutext.Lock()

	if p.Conn != nil {
		p.Conn.Close()
	}
	close(p.PieceDone)
	p.ConnMutext.Unlock()
	close(p.PieceWorkRecvChan)


	for ;; {
		val, ok := <- p.PieceWorkRecvChan
		if !ok{
			break
		}
		p.FailledPiece <- val
	}
}

func (p *Peer) SetBitField(bitfield []byte) {
	if !p.IsGood() {
		return
	}
	p.BitField = make([]byte, len(bitfield))
	copy(p.BitField, bitfield)
}

func (p *Peer) InitBitField(size int) {

	p.BitField = make([]byte, size)
	//copy(p.BitField, bitfield)
}

