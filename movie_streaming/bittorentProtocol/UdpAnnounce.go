package bittorentProtocol

import (
	"bytes"
	"encoding/binary"
)

type AnnounceRequest struct {
	ConnectionId  int64
	Action        TrackerAction
	TransactionId int32
	InfoHash      [20]byte
	PeerId        [20]byte
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         int32
	IP            int32
	Key           int32
	NumWhat       int32
	Port          uint16
}

func (a *AnnounceRequest) Serialize() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, a)
	return buf.Bytes()
}

type AnnounceResponse struct {
	Action        TrackerAction
	TransactionId int32
	Interval      int32
	Seeders       int32
	Leechers      int32
}

func NewAnnounceResponse(resp []byte) *AnnounceResponse {

	var (
		connResp AnnounceResponse
	)

	buf := bytes.NewBuffer(resp[:20])
	binary.Read(buf, binary.BigEndian, &connResp)

	return &connResp
}
