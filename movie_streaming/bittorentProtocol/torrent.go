package bittorentProtocol

import (
	bencode "github.com/jackpal/bencode-go"
	"io"
)

type TrackerAction int32

// 2. Declare the protocol actions using iota
const (
	ActionConnect  TrackerAction = iota // 0
	ActionAnnounce                      // 1
	ActionScrape                        // 2
	ActionError                         // 3
)

type FileInfo struct {
	Path   []string `bencode:"path"`
	Length int64    `bencode:"length"`
}

type InfoMultipleFiles struct {
	Files       []FileInfo `bencode:"files"`
	PieceLength int64      `bencode:"piece length"`
	Pieces      string     `bencode:"pieces"`
	Name        string     `bencode:"name"`
}

type TorrentFile struct {
	Announce     string            `bencode:"announce"`
	AnnounceList [][]string        `bencode:"announce-list"`
	Info         InfoMultipleFiles `bencode:"info"`
	InfoHash     []byte
}

func NewTorrent(r io.Reader) (*TorrentFile, error) {
	var (
		t TorrentFile
	)
	err := bencode.Unmarshal(r, &t)

	return &t, err
}

func (t *TorrentFile) CalculateLength() int64 {
	res := int64(0)
	for _, val := range t.Info.Files {
		res += val.Length
	}
	return res

}
