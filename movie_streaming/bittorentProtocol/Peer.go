package bittorentProtocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

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
	valid       bool
	UnChoke     bool
	PacketSent  int
	ChokeUpload bool

	InfoHash [20]byte

	PeerMutex    sync.Mutex
	ConnMutext   sync.Mutex
	MessagesSent int
	ClientID     [20]byte

	PieceWorkRecvChan chan *PieceWork
	PieceWorkResChan  chan *PieceWork
	Started           bool
	BroadCastMsg      chan []byte

	CurPiece     *PieceWork
	Interested   bool
	BytesRecived uint32
	fileUpload   []*FileUploads
	HasBitField  bool

	ServerBitField []byte
}

func (p *Peer) SetUpServerBitField(b []byte) {
	p.ServerBitField = b
}

func (p *Peer) SetUpFileUploads(fileUpload []*FileUploads) {
	p.fileUpload = fileUpload
}

func (p *Peer) RecivedBytes(n uint32) {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	p.BytesRecived += n
}

func (p *Peer) ResetBytesReceived() {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	p.BytesRecived = 0
}

func (p *Peer) UnChokePeer(status bool) {

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
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	return p.ChokeUpload
}

func (p *Peer) UpdateChokeUpload(status bool) {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	p.ChokeUpload = status
}

func (p *Peer) GetBytesReceived() uint32 {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	return p.BytesRecived
}

func (p *Peer) InterestedStatus(status bool) {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	p.Interested = status
}

func (p *Peer) IsInterested() bool {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	return p.Interested
}

func (p *Peer) SetChannel(PieceWorkRecvChan chan *PieceWork, PieceWorkResChan chan *PieceWork) {
	p.PieceWorkRecvChan = PieceWorkRecvChan
	p.PieceWorkResChan = PieceWorkResChan
}

func (p *Peer) SentMessage() {
	p.PeerMutex.Lock()
	p.MessagesSent += 1
	p.PeerMutex.Unlock()
}

func (p *Peer) ReceivedMessage() {
	p.PeerMutex.Lock()
	p.MessagesSent -= 1
	if p.MessagesSent < 0 {
		p.MessagesSent = 0
	}
	p.PeerMutex.Unlock()
}

func (p *Peer) CanSendMessage() bool {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()

	return p.MessagesSent < 32
}

func (p *Peer) GetMessageSent() int {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	return p.MessagesSent
}

func (p *Peer) SetGood(b bool) {
	p.PeerMutex.Lock()
	p.valid = b
	p.PeerMutex.Unlock()
}

func (p *Peer) IsGood() bool {
	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()
	return p.valid
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

		peers = append(peers, &Peer{IP: ip, Port: port, valid: true})
	}

	return peers
}

func (p *Peer) Connect() {
	conn, err := net.DialTimeout("tcp", p.IP.String()+":"+strconv.FormatUint(uint64(p.Port), 10), 5*time.Second)
	if err != nil {
		p.SetGood(false)
		return
	}
	p.valid = true
	p.Conn = conn
}

func (p *Peer) PeerHandShake(h *HandShake) []byte {

	if !p.IsGood() {
		return nil
	}

	rawBytes := h.Serialize()

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
	if _, err := p.Conn.Write(rawBytes); err != nil {
		p.SetGood(false)
		return nil
	}
	resp := make([]byte, 68)

	p.Conn.SetReadDeadline(time.Now().Add(time.Second * 15))
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

	p.Conn.SetDeadline(time.Now().Add(time.Second * 15))
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
	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
	m := Msg{ID: MsgInterested}
	data := m.Serialize()
	if _, err := p.Conn.Write(data); err != nil {
		p.SetGood(false)
		return err
	}

	return nil
}

func (p *Peer) Request(cur uint32, base uint32, length uint32) error {
	if !p.IsGood() || !p.UnChoke {
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

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
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

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
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
		p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
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

	p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
	if _, err := p.Conn.Write(msg.Serialize()); err != nil {
		p.SetGood(false)
		return err
	}

	fmt.Println("----------------------------------- sent bif filed", p.ID())
	return nil
}

func (p *Peer) PeerGoRotine(wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		other sync.WaitGroup
	)
	defer other.Wait()

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
			p.PieceWorkRecvChan <- piece
			return
		}

		if !p.HasPiece(piece.Index) || !p.PeerIsChoking() {
			p.PieceWorkRecvChan <- piece
			time.Sleep(100 * time.Millisecond)
			continue
		}
		piecesChan := make(chan uint32, piece.NPiece+1)

		for begin := uint32(0); begin < piece.Size; begin += piece.BlockSize {
			piecesChan <- begin
		}

		p.SetCurrentPiece(piece)
		timeoutTicker := time.NewTicker(3 * time.Second)
		defer timeoutTicker.Stop()

		limit := time.Now().Add(5 * time.Second)
		lastDownload := piece.GetDownloaded()
		for !piece.Done() && p.IsGood() {

			select {
			case begin := <-piecesChan:
				if piece.IsThisDone(begin) {
					continue
				}
				if !p.CanSendMessage() || !p.PeerIsChoking() {
					piecesChan <- begin
					time.Sleep(100 * time.Millisecond)
					continue
				}

				length := piece.BlockSize
				if begin+length > piece.Size {
					length = piece.Size - begin
				}
				p.Request(piece.Index, begin, length)
				p.SentMessage()
			case <-timeoutTicker.C:
				piece.PrintState()
				if time.Until(limit) < 0 {
					if lastDownload == piece.GetDownloaded() {
						p.SetGood(false)
						break
					} else {
						limit = time.Now().Add(5 * time.Second)
						lastDownload = piece.GetDownloaded()
					}
				}
			}
		}
		if piece.Done() {
			p.PieceWorkResChan <- piece
		} else {
			p.SetCurrentPiece(nil)
			close(piecesChan)
			fmt.Println("peer time outed", p.ID())
			p.Clear()
			p.PieceWorkRecvChan <- piece
			return
		}
		p.SetCurrentPiece(nil)
	}
}

func (p *Peer) PeerChokingState(state bool) {

	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()

	p.UnChoke = state
}

func (p *Peer) PeerIsChoking() bool {

	p.PeerMutex.Lock()
	defer p.PeerMutex.Unlock()

	return p.UnChoke
}

func (p *Peer) PeerMesgs(wg *sync.WaitGroup) {
	defer wg.Done()
	if !p.IsGood() {
		return
	}
	for {
		if !p.IsGood() {
			return
		}
		m, err := p.GetMessage()
		if err != nil {
			p.SetGood(false)
			return
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
				p.SetBitField(m.Payload)
			}
		case MsgPiece:
			p.ReceivedMessage()
			index, begin, buf, ok := m.ParsePiece()
			// fmt.Println("Piece", index, "was received from ", p.ID())
			if !ok {
				continue
			}
			if p.CurPiece != nil {
				p.CurPiece.SetPiece(index, buf, begin)
				p.RecivedBytes(uint32(len(buf)))
			}
		case MsgRequest:
			fmt.Println("--------------------received a messag request --------------------")
			if !p.IsInterested() || p.ChokeUploadStatus() {
				continue
			}
			index, begin, length, ok := m.ParseRequest()
			if !ok {
				fmt.Println("request isn't good")
				continue
			}
			piece, err := GetCurrentPiece(index, begin, length, p.fileUpload)
			if err != nil {
				continue
			}
			if err := p.SendPiece(index, begin, piece); err != nil {
				fmt.Println("error sending piece")
			}
			// here we need to send the piece

		case MsgInterested:
			fmt.Println("-------------------- Peer is interested")
			p.InterestedStatus(true)
		case MsgNotInterested:
			fmt.Println("--------------------- Peer is not interested")
			p.InterestedStatus(false)
		}
	}
	// here we will be waiting for peer messages
}

// Clear function to free all resources allocated
func (p *Peer) Clear() {
	p.PeerMutex.Lock()
	p.valid = false
	// close(p.PieceWorkRecvChan)
	// close(p.PieceWorkResChan)
	if p.Conn != nil {
		p.Conn.Close()
	}
	p.PeerMutex.Unlock()
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
