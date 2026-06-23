package types 


import (
    "fmt"
    
    "io"
    "strconv"
    "net/url"
    "strings"
    bencode "github.com/jackpal/bencode-go"
)




type FileInfo struct {
    Path []string `bencode:"path"`
    Length int `bencode:"length"`
} 

type InfoMultipleFiles struct {
    Files []FileInfo `bencode:"files"`
    PieceLength int `bencode:"piece length"`
    Pieces string `bencode:"pieces"`
    Name string  `bencode:"name"`
}


type TorrentFile struct {
    Announce string `bencode:"announce"`
    AnnounceList [][]string `bencode:"announce-list"`
    Info  InfoMultipleFiles  `bencode:"info"`
    InfoHash []byte
}




func NewTorrent(r io.Reader) (*TorrentFile, error) {
    var (
        t TorrentFile
    )
    err := bencode.Unmarshal(r,&t)



    return  &t, err
}


func (t *TorrentFile) CalculateLength() int{
    res := 0
    for _, val := range(t.Info.Files) {
        res += val.Length
    }
    return  res 

}

func (t *TorrentFile) BuildTrackerUrl(peerID [20]byte, port string) []string{
     var (
        urls []string
        res []string
     )


    if strings.HasPrefix(t.Announce, "http:") || strings.HasPrefix(t.Announce, "https:"){
                urls = append(urls, t.Announce)
    }

     for _ ,values := range(t.AnnounceList){

         for _, val := range(values) {

             if strings.HasPrefix(val, "http:") || strings.HasPrefix(val, "https:") {
                 urls = append(urls, val)
             }
         }
     }


    for _, u := range(urls) {
        base, _ := url.Parse(u)

        params := url.Values {
        "info_hash":  []string{string(t.InfoHash[:])},
        "peer_id":    []string{string(peerID[:])},
        "port":       []string{port},
        "uploaded":   []string{"0"},
        "downloaded": []string{"0"},
        "compact":    []string{"1"},
        "left":       []string{strconv.Itoa(t.CalculateLength())},
        }
        base.RawQuery = params.Encode()
        res = append(res, base.String())
    }

    for _, val := range(res) {
        fmt.Println(val)
    }

    return nil
}
