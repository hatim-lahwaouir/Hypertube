package bittorentProtocol

import (
	"bytes"
	"crypto/sha1"
	"io"


	bencode "github.com/jackpal/bencode-go"
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
	Length      int64      `bencode:"length,omitempty"`
}

type TorrentFile struct {
	Announce     string            `bencode:"announce"`
	AnnounceList [][]string        `bencode:"announce-list"`
	Info         InfoMultipleFiles `bencode:"info"`
	Name 		 string     				`bencode:"name"`
	//   "length": 3808117223,
    //   "name": "Spider-Man- Brand New Day 2026.1080p.HQ Pre.Multi.AAC 2.0.x264.mkv",
	InfoHash     []byte
}

func NewTorrent(r io.Reader) (*TorrentFile, error) {
    var t TorrentFile
    
    err := bencode.Unmarshal(r, &t)
    if err != nil {
        return nil, err
    }

    var buf bytes.Buffer
    err = bencode.Marshal(&buf, t.Info)
    if err != nil {
        return nil, err
    }

    hash := sha1.Sum(buf.Bytes())
    
    t.InfoHash = hash[:]

	if  t.Info.Length  != 0{
			t.Info.Files = append(t.Info.Files, FileInfo{Path: []string{t.Info.Name}, Length: t.Info.Length })
	}
    return &t, nil
}

func (t *TorrentFile) CalculateLength() int64 {
	res := int64(0)
	for _, val := range t.Info.Files {
		res += val.Length
	}
	return res

}
