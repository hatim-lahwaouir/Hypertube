package types 


import (
    "io"
    bencode "github.com/jackpal/bencode-go"
)




type FileInfo struct {
    Path []string `bencode:"path"`
    Length int `bencode:"length"`


} 

type InfoMultipleFiles struct {
    Files []FileInfo `bencode:"files"`
    PieceLength int `bencode:"piece length"`
    Pieces int `bencode:"pieces"`
    Name string  `bencode:"name"`
}


type TorrentParse struct {
    Announce string `bencode:"announce-list"`
    AnnounceList [][]string `bencode:"announce-list"`
}




func NewTorrent(r io.Reader) (*TorrentParse, error) {
    var (
        t TorrentParse
    )
    err := bencode.Unmarshal(r,&t)

    return  &t, err
}
