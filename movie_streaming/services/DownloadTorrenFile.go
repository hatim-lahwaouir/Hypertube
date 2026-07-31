package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/bittorentProtocol"
	"golang.org/x/text/unicode/rangetable"
)





type  DownloadTorrent struct {
	magnetLink string
}



func NewDownloadTorrent(m string) *DownloadTorrent{
	return &DownloadTorrent{magnetLink:  m}
}





func (d *DownloadTorrent) DownloadTorrent() error {
	var( 
		udpTrackers []*bittorentProtocol.UdpTracker 
		peerId [20]byte
		Peers []*bittorentProtocol.Peer
	)


	if _, err := rand.Read(peerId[:]); err != nil {
		return err
	}

	magnetLink, err := url.Parse(d.magnetLink)
	if err != nil {
		return err
	}

	for _,val := range magnetLink.Query()["tr"]{
		
		if strings.HasPrefix(val, "udp"){
			udpTrackers = append(udpTrackers, bittorentProtocol.NewUdpTracker(val))
		}
	}
	for i := range(udpTrackers){
		fmt.Println(udpTrackers[i])
	}

	InfoHash, err := hex.DecodeString(magnetLink.Query()["xt"][:][0][len("urn:btih:"):])
	if err != nil {
		return err
	}
	
	source := bittorentProtocol.NewUdpTrackers(udpTrackers)
	Peers = source.GetPeers(&bittorentProtocol.CurrentState{InfoHash: [20]byte(InfoHash), PeerId: peerId})


	// we should connect to each peer
	// send them the handshake and wait if they support the extension

	handShake := bittorentProtocol.NewHandShake(peerId, [20]byte(InfoHash))



	

	return nil
}