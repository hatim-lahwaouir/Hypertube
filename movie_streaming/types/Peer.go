package types

import (
	"encoding/binary"
	"time"
    "fmt"
    "strconv"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
    Conn  net.Conn
    IsGood bool
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

		peers = append(peers, Peer{IP: ip, Port: port})
	}

	return peers
}



func (p *Peer) Connect() {
    conn , err := net.DialTimeout("tcp", p.IP.String() + ":" +  strconv.FormatUint(uint64(p.Port), 10), 1 * time.Second)
    if err != nil {
        //fmt.Println(err)
        p.IsGood = false
        return
    }
    if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		fmt.Println("errror setting dead line  ", err.Error())
		return 
	}

    p.IsGood = true 
    p.Conn = conn 
}



func (p *Peer) PeerHandShake(h HandShake) {

    if p.IsGood == false{
        return
    }
    start := time.Now()


        
    rawBytes := h.Serialize()
    
    p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 1))
    if _, err := p.Conn.Write(rawBytes); err != nil {
            fmt.Println("error sending data", err)
            p.IsGood  = false
            return 
    }
    resp := make([]byte, 68)


    p.Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
    n , err := p.Conn.Read(resp)
    if err != nil {
            p.IsGood  = false
            return 
    }
    fmt.Println("handshake ->", n, "in", start.Sub(time.Now()))

}

