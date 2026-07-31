package bittorentProtocol

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"net"
	"net/url"
	"sync"
	"time"
)

type UdpTracker struct {
	Scheme       string
	Host         string
	ConnectionId int64
	IsGood       bool
	lastTime    time.Time
}


type UdpTrackers struct {
	trackers []*UdpTracker
	peerResult []*Peer
}



func NewUdpTrackers(UdpTracker []*UdpTracker) *UdpTrackers {

	return &UdpTrackers{trackers: UdpTracker}
}


func NewUdpTracker(trackerUrl string) *UdpTracker {
	endpoint, _ := url.Parse(trackerUrl)

	return &UdpTracker{Scheme: endpoint.Scheme, Host: endpoint.Host}
}



func (u *UdpTrackers) GetPeers(cur *CurrentState) []*Peer {
    var (

        waitGetUdpPeers sync.WaitGroup
        waitWorkersUdpPeers sync.WaitGroup
        peerRes chan []*Peer
        recv chan *UdpTracker
        n_gorotines int
    )
	u.peerResult = nil
    n_gorotines = 20

	
    recv = make(chan *UdpTracker, 50)
    peerRes = make(chan []*Peer, 50)



	// lanch gorotines that each one will get some peers from udp trackesr and store them in peerRes chanel 
	for i := 0; i < n_gorotines; i += 1 {
        waitWorkersUdpPeers.Add(1)
        go getUdpPeersWorker(recv, peerRes,cur , &waitWorkersUdpPeers)
    }
    waitGetUdpPeers.Add(1)
    go u.storePeersFromUdpTracker(peerRes, &waitGetUdpPeers)
	for _, udptracker := range u.trackers {
        recv <- udptracker
	}
    close(recv)
    waitWorkersUdpPeers.Wait()
    close(peerRes)
    waitGetUdpPeers.Wait()

	return u.peerResult
}


func getUdpPeersWorker(recv chan *UdpTracker, res chan []*Peer,  cur *CurrentState ,wg *sync.WaitGroup) {
   

   defer wg.Done()

   for  tr := range(recv) {
        tr.GetConnectionId()
        p := tr.GetPeers(cur)
        res <- p
   }
}

func (u *UdpTrackers) storePeersFromUdpTracker(res chan []*Peer, wg *sync.WaitGroup) {

	defer wg.Done()
    
    for p := range(res) {
        u.peerResult = append(u.peerResult, p...)
    }
}



func (u *UdpTracker) GetConnectionId() {
	var (
		buf      *bytes.Buffer
		resp     []byte
		connResp ConnectionResp
	)
	if u.lastTime.IsZero(){
		u.lastTime = time.Now()
	}
	if time.Until(u.lastTime.Add(2 * time.Minute)) < 0 {
		return
	}
	u.lastTime = time.Now()
	
	conn, err := net.DialTimeout(u.Scheme, u.Host, 3*time.Second)

	if err != nil {
		return
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return
	}

	req, rawReq := NewConnectionReq()
	_, err = conn.Write(rawReq)
	if err != nil {
		return
	}
	resp = make([]byte, 16)
	n, err := conn.Read(resp)

	if err != nil {
		return
	}

	if n < 16 {
		return
	}

	buf = bytes.NewBuffer(resp)
	err = binary.Read(buf, binary.BigEndian, &connResp)

	if err != nil {
		return
	}
	if connResp.TransactionId != req.TransactionId {
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

func (u *UdpTracker) GetPeers(state *CurrentState) []*Peer {
	var (
		resp []byte
	)
	if !u.IsGood  {
		return nil
	}

	conn, err := net.DialTimeout(u.Scheme, u.Host, time.Second*1)
	if err != nil {
		return nil
	}
	defer conn.Close()

	req := u.NewAnnounceRequest(state)
	//binary.Write(buf, binary.BigEndian, &req)
	_, err = conn.Write(req.Serialize())
	if err != nil {
		return nil
	}
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return nil
	}

	resp = make([]byte, 3000)
	n, err := conn.Read(resp)
	if err != nil {
		return nil
	}

	if n < 20 {
		return nil
	}

	connResp := NewAnnounceResponse(resp)
	if err != nil {
		return nil
	}

	if req.TransactionId != connResp.TransactionId || connResp.Action != ActionAnnounce {
		return nil
	}

	return NewPeers(resp, n)
}
