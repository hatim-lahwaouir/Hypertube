package bittorentProtocol

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

type ConnectionReq struct {
	ProtocolId    int64
	Action        TrackerAction
	TransactionId int32
}

func NewConnectionReq() (ConnectionReq, []byte) {
	var trId int32
	binary.Read(rand.Reader, binary.BigEndian, &trId)
	req := ConnectionReq{
		ProtocolId:    int64(0x41727101980),
		Action:        ActionConnect,
		TransactionId: trId,
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, &req)
	return req, buf.Bytes()
}

type ConnectionResp struct {
	Action        TrackerAction
	TransactionId int32
	ConnectionId  int64
}
