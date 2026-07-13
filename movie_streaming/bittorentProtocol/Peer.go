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
	IP   net.IP
	Port uint16
    Conn  net.Conn
    BitField []byte
    valid bool
    UnChoke bool
    PacketSent  int

    InfoHash [20]byte
    Mu      sync.RWMutex
    PeerMutex      sync.Mutex
    MessagesSent  int
    ClientID [20]byte

    PieceWorkRecvChan chan *PieceWork
    PieceWorkResChan chan *PieceWork
    Started bool


    CurPiece *PieceWork
}


func (p *Peer)SentMessage(){
    p.PeerMutex.Lock()
    p.MessagesSent += 1
    p.PeerMutex.Unlock()
}

func (p *Peer) ReceivedMessage(){
    p.PeerMutex.Lock()
    p.MessagesSent -= 1
    if p.MessagesSent < 0{
        p.MessagesSent = 0
    }
    p.PeerMutex.Unlock()
}


func (p *Peer) CanSendMessage() bool {
    p.PeerMutex.Lock()
    defer p.PeerMutex.Unlock()
    
    return p.MessagesSent < 4
}

func (p *Peer) GetMessageSent () int {
    p.PeerMutex.Lock()
    defer p.PeerMutex.Unlock()
    return p.MessagesSent
}

func (p *Peer) SetGood(b bool) {
    p.Mu.Lock()
    p.valid = b
    p.Mu.Unlock()
}


func (p *Peer) IsGood() bool {
    p.Mu.RLock()
    
    defer p.Mu.RUnlock()
    return p.valid
}
 
func (p *Peer) Id() string {
    return p.IP.String() + ":" +  strconv.FormatUint(uint64(p.Port), 10)

}


func (p *Peer) SetInfo(infoHash [20]byte, clientID [20]byte)  {
    p.InfoHash = infoHash
    p.ClientID = clientID
}


func (p *Peer)HasPiece(index uint32) bool {
      p.PeerMutex.Lock()
    defer p.PeerMutex.Unlock()
    if p.BitField == nil{
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


func NewPeers(resp []byte, n int) []Peer {
	var peers []Peer

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

		peers = append(peers, Peer{IP: ip, Port: port,     PieceWorkRecvChan : make(chan *PieceWork, 5),
        PieceWorkResChan : make(chan *PieceWork, 5),  valid: true, })
	}

	return peers
}



func (p *Peer) Connect() {
    conn , err := net.DialTimeout("tcp", p.IP.String() + ":" +  strconv.FormatUint(uint64(p.Port), 10), 2 * time.Second)
    if err != nil {
        p.SetGood(false)
        return
    }
    p.valid = true
    p.Conn = conn 
}



func (p *Peer) PeerHandShake(h *HandShake) []byte {

    if !p.IsGood(){
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
    _ , err := p.Conn.Read(resp)
    if err != nil {
            p.SetGood(false)
            return nil
    }

    return  resp
}

func (p *Peer) ValidHandShake(h *HandShake, peerResp []byte ) bool {

    if !p.IsGood(){
        return false
    }
    rawBytes := h.Serialize()
     // comapre hash info and pstr
    if !bytes.Equal(peerResp[28:len(peerResp) - 20], rawBytes[28:len(rawBytes) - 20]) ||
       ! bytes.Equal(rawBytes[0:20], peerResp[0:20])  {
        p.SetGood(false)
        return  false
    }



    return true
}

func (p *Peer) GetMessage() (*Msg, error) {

    if !p.IsGood(){
        return nil, nil 
    }

    p.Conn.SetDeadline(time.Now().Add(time.Second * 15))
    m, err := NewMessage(p.Conn)
    if err != nil {
        if errors.Is(err, os.ErrDeadlineExceeded) {
			return nil,nil
		}
        p.SetGood(false)
        return nil,err
    }
    return m, nil
}



func (p *Peer) Intersted() (error) {
    if !p.IsGood(){
        return nil 
    }

    p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
    m := Msg{ID: MsgInterested}
    data := m.Serialize()
    if _, err := p.Conn.Write(data); err != nil {
            p.SetGood(false) 
            return  err
    }
    
    return nil
}

func (p *Peer) Request(cur uint32 ,base uint32,length uint32) error {
    if !p.IsGood() || !p.UnChoke {
        return errors.New("invalid Peer") 
    }
    buf := make([]byte, 17)

	binary.BigEndian.PutUint32(buf[0:4], 13)

	buf[4] = byte(MsgRequest)

	binary.BigEndian.PutUint32(buf[5:9],cur )

	binary.BigEndian.PutUint32(buf[9:13], base)

	binary.BigEndian.PutUint32(buf[13:17], length)

    p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 15))
    if _, err := p.Conn.Write(buf); err != nil {
            p.SetGood(false)
            return   err
    }
    p.PacketSent++
    return nil

}

func (p *Peer) PeerGoRotine(wg *sync.WaitGroup)  {
    defer wg.Done()
    var (
        other sync.WaitGroup
    )

    p.Connect() 
    if !p.IsGood(){
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

    fmt.Println("hello", p.Id())
    // here we will start a gorotine for reading peer messages 
    other.Add(1)
    p.Intersted()

    go  p.PeerMesgs(&other)

    for piece := range(p.PieceWorkRecvChan){
        piecesChan := make(chan uint32, piece.NPiece + 1)

        for begin := uint32(0); begin < piece.Size; begin += piece.BlockSize {
            piecesChan <- begin
        }

        p.PeerMutex.Lock()
        p.CurPiece = piece
        p.PeerMutex.Unlock()



        for !piece.Done(){
            for begin := range(piecesChan) {
                if piece.IsThisDone(begin){
                    if piece.Done() {
                        break
                    }
                    continue
                }
                if !p.PeerIsChoking() || !p.CanSendMessage() || !p.HasPiece(piece.Index) {
                    piecesChan <- begin
                    time.Sleep(100 * time.Millisecond)
                    continue
                }

                length := piece.BlockSize
                if begin + length > piece.Size {
                    length = piece.Size - begin 
                }
                // fmt.Println("Requeest for piece  " ,piece.Index, begin / piece.BlockSize, "was sent to ", p.Id())
                p.Request(piece.Index, begin, length)
                p.SentMessage()
            
                if piece.Done() {
                        break
                }
            } 
        }
    }
    other.Wait()
}

func (p *Peer) PeerChokingState(state bool){

    p.PeerMutex.Lock()
    defer p.PeerMutex.Unlock()

    p.UnChoke = state
}

func (p *Peer) PeerIsChoking() bool{

    p.PeerMutex.Lock()
    defer p.PeerMutex.Unlock()

    return p.UnChoke
}


func (p *Peer) PeerMesgs(wg *sync.WaitGroup)  {
    defer wg.Done()
    if !p.IsGood(){
        return
    }
    for ;; {
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
            case MsgBitfield :
                if len(p.BitField) == len(m.Payload)  {
                    p.SetBitField(m.Payload)
                }
            case MsgPiece:
                p.ReceivedMessage()
                index, begin, buf, ok := m.ParsePiece()
                // fmt.Println("Piece", index, "was received from ", p.Id())
                if !ok{
                    continue
                }
                p.PeerMutex.Lock()
                if p.CurPiece != nil {
                    p.CurPiece.SetPiece(index, buf, begin)
                }
                p.PeerMutex.Unlock()
        }
    }
    // here we will be waiting for peer messages
}


// Clear function to free all resources allocated 
func (p *Peer) Clear() {
    p.Mu.Lock()
    p.valid = false
    close(p.PieceWorkRecvChan)
    close(p.PieceWorkResChan)
    if p.Conn != nil {
        p.Conn.Close()
    }
    p.Mu.Unlock()
}



func (p *Peer) SetBitField(bitfield []byte) {
    if !p.IsGood(){
        return
    }
    p.BitField = make([]byte, len(bitfield))
    copy(p.BitField, bitfield)
}

func (p *Peer) InitBitField(size int) {

    p.BitField = make([]byte,size)
    //copy(p.BitField, bitfield)
}

