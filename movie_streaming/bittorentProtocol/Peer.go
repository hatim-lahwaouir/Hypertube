package bittorentProtocol 

import (
	"encoding/binary"
    "bytes"
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
    conn , err := net.DialTimeout("tcp", p.IP.String() + ":" +  strconv.FormatUint(uint64(p.Port), 10), 2 * time.Second)
    if err != nil {
        //fmt.Println(err)
        p.IsGood = false
        return
    }
    if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		fmt.Println("errror setting dead line  ", err.Error())
		return 
	}

    p.IsGood = true 
    p.Conn = conn 
}



func (p *Peer) PeerHandShake(h HandShake) []byte {

    if p.IsGood == false{
        return nil
    }
    start := time.Now()


        
    rawBytes := h.Serialize()
    
    p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
    if _, err := p.Conn.Write(rawBytes); err != nil {
            fmt.Println("error sending data", err)
            p.IsGood  = false
            return nil
    }
    resp := make([]byte, 68)


    p.Conn.SetReadDeadline(time.Now().Add(time.Second * 3))
    n , err := p.Conn.Read(resp)
    if err != nil {
            p.IsGood  = false
            return nil
    }
    fmt.Println("handshake ->", n, "in", start.Sub(time.Now()))

    return  resp
}

func (p *Peer) ValidHandShake(h HandShake, peerResp []byte ) bool {

    if p.IsGood == false{
        return false
    }
    rawBytes := h.Serialize()


    


     // comapre hash info and pstr

    if bytes.Equal(peerResp[28:len(peerResp) - 20], rawBytes[28:len(rawBytes) - 20]) == false  ||
        bytes.Equal(rawBytes[0:20], peerResp[0:20]) == false {

    
        

       p.IsGood  = false
        return  false
    }




    return true
}

