package services

import (
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "crypto/rand"
    "fmt"
    "strconv"
    "net/url"
    "path/filepath"
    "strings"
)


type TorrentStreaming struct {
    PeerId [20]byte
    InfoHash []byte
    Port    string
    Length  int
    PieceLength int
    Pieces [][20]byte
    
    Files []string
    Size  []int


    HttpTrackers []string
    UdpTrackers  []string

    Downloaded int
    Left       int
}


type CurrentState struct {
    Downloaded int
    Left       int
}






func NewTorrentStreaming(p string, t *types.TorrentFile) *TorrentStreaming {

    var (
        files []string
        size  []int
        pieces [][20]byte
        httpTrackers []string
        udpTrackers []string
        peerId [20]byte
    )


    for _, val := range(t.Info.Files) {
        files = append(files,filepath.Join(val.Path...))
        size = append(size ,val.Length)
    }
    

    hashSize := 20
    piecesByte :=  []byte(t.Info.Pieces)
    numPieces := len(piecesByte) / hashSize


    for i := 0; i < numPieces; i +=1{

        start := i * hashSize
        end   := start + hashSize

        b := [20]byte(piecesByte[start: end])
        pieces = append(pieces, b)
    }



    if strings.HasPrefix(t.Announce, "http:") || strings.HasPrefix(t.Announce, "https:"){
                httpTrackers = append(httpTrackers, t.Announce)
    }else {

                udpTrackers = append(udpTrackers, t.Announce)
    }

     for _ ,values := range(t.AnnounceList){

         for _, val := range(values) {

             if strings.HasPrefix(val, "http:") || strings.HasPrefix(val, "https:") {
                 httpTrackers = append(httpTrackers, val)
             } else {

                 udpTrackers = append(udpTrackers, val)
             }
         }
     }


     if _, err :=  rand.Read(peerId[:]); err != nil {
            peerId = [20]byte{72, 101, 108, 108, 111}
     }

    return &TorrentStreaming{PeerId: peerId, InfoHash: t.InfoHash, Port: p, Length: t.CalculateLength(),
        Size: size, Files: files, UdpTrackers: udpTrackers, HttpTrackers: httpTrackers, Left:  t.CalculateLength(), Downloaded: 0,}
}




func (t *TorrentStreaming) GetPeers(cur *CurrentState) {
    for _, u := range(t.HttpTrackers) {

        base, _ := url.Parse(u)

        params := url.Values {
        "info_hash":  []string{string(t.InfoHash[:])},
        "peer_id":    []string{string(t.PeerId[:])},
        "port":       []string{t.Port},
        "uploaded":   []string{"0"},
        "downloaded": []string{strconv.Itoa(cur.Downloaded)},
        "compact":    []string{"1"},
        "left":       []string{strconv.Itoa(cur.Left)},
        }
        base.RawQuery = params.Encode()
        fmt.Println(base)
    }

    fmt.Println(t.UdpTrackers, len(t.UdpTrackers) )
    fmt.Printf("%x", t.PeerId)
}








