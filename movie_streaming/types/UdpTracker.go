package types

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"time"
)

type UdpTracker struct {
	Scheme       string
	Host         string
	ConnectionId int64
	IsGood       bool
}



func NewUdpTracker(trackerUrl string) UdpTracker {
	endpoint, _ := url.Parse(trackerUrl)

	return UdpTracker{Scheme: endpoint.Scheme, Host: endpoint.Host}
}

func (u *UdpTracker) GetConnectionId() {
	var (
		buf      *bytes.Buffer
		resp     []byte
		connResp ConnectionResp
	)

	conn, err := net.DialTimeout(u.Scheme, u.Host, 1 * time.Second)

	if err != nil {
		fmt.Println("errror starting connection", err.Error())
		return
	}
	defer conn.Close()
    if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		fmt.Println("errror setting dead line  ", err.Error())
		return 
	}

	req, rawReq := NewConnectionReq()
	_, err = conn.Write(rawReq)
	if err != nil {
		fmt.Println("errror writing to the packet ", err.Error())
		return
	}
	resp = make([]byte, 16)
	n, err := conn.Read(resp)

	if err != nil {
		fmt.Println("errror reading from connection  ", err.Error())
		return
	}

	if n < 16 {
		fmt.Println("not all data was provided", err.Error())
		return
	}

	buf = bytes.NewBuffer(resp)
	err = binary.Read(buf, binary.BigEndian, &connResp)

	if err != nil {
		fmt.Println("error reading response to the struct ", err.Error())
		return
	}
	if connResp.TransactionId != req.TransactionId {
		fmt.Println("error reading response to the struct ", err.Error())
		return
	}
	if connResp.TransactionId == req.TransactionId {
		u.ConnectionId = connResp.ConnectionId
		u.IsGood = true
	}
}

func (u *UdpTracker) NewAnnounceRequest(state *CurrentState) *AnnounceRequest {

	var trId int32
	binary.Read(rand.Reader, binary.BigEndian, &trId)
	return &AnnounceRequest{
		ConnectionId:  u.ConnectionId,
		Action:        ActionAnnounce, // announce
		TransactionId: trId,
		InfoHash:      state.InfoHash,
		PeerId:        state.PeerId,
		Downloaded:    state.Downloaded,
		Left:          state.Left,
		NumWhat:       -1,
		Uploaded:      0,
		Port:          state.Port,
	}
}

func (u *UdpTracker) GetPeers(state *CurrentState) ([]Peer) {
	var (
		resp []byte
	)
	if u.IsGood == false {
		return nil
	}

	conn, err := net.DialTimeout(u.Scheme, u.Host, time.Second * 1)
	if err != nil {
		fmt.Println("errror starting connection", err.Error())
		return nil
	}
	defer conn.Close()

	req := u.NewAnnounceRequest(state)
	//binary.Write(buf, binary.BigEndian, &req)
	_, err = conn.Write(req.Serialize())
	if err != nil {
		fmt.Println("errror writing to the packet ", err.Error())
		return nil
	}
	if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		fmt.Println("errror setting dead line  ", err.Error())
		return nil
	}

	resp = make([]byte, 3000)
	n, err := conn.Read(resp)
	if err != nil {
		fmt.Println("errror reading response ", err.Error())
		return nil
	}

	if n < 20 {
		fmt.Println("errror reading response ", err.Error())
		return nil
	}

	connResp := NewAnnounceResponse(resp)
	if err != nil {
		fmt.Println("errror parrsing response ", err.Error())
		return nil
	}

	if req.TransactionId != connResp.TransactionId || connResp.Action != ActionAnnounce {
        return nil
	}

    return NewPeers(resp, n)
}
