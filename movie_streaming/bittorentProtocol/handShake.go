package bittorentProtocol

import (
	"bytes"
)

type HandShake struct {
	PeerId   [20]byte
	InfoHash [20]byte
}

func NewHandShake(peerId [20]byte, infoHash [20]byte) *HandShake {

	return &HandShake{PeerId: peerId, InfoHash: infoHash}
}

func (h *HandShake) Serialize() []byte {
	var (
		buff *bytes.Buffer
	)
	buff = new(bytes.Buffer)

	buff.Write([]byte{19})
	buff.Write([]byte("BitTorrent protocol"))
	buff.Write(bytes.Repeat([]byte{0}, 8))
	buff.Write(h.InfoHash[:])
	buff.Write(h.PeerId[:])

	return buff.Bytes()
}
