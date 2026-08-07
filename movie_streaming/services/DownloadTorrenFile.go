package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/anacrolix/torrent"
)





type  DownloadTorrent struct {
	magnetLink string
	maxConn chan bool
}



func NewDownloadTorrent(m string) *DownloadTorrent{
	return &DownloadTorrent{magnetLink:  m, maxConn: make(chan bool, 5)}
}





func (d *DownloadTorrent) DownloadTorrent() (*os.File, error) {

	d.maxConn <- true
	file, err := os.CreateTemp("","file.*.torrent")
	if err != nil {
		return nil, err
	}

	

	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0 
	cfg.DataDir = os.TempDir()


	// 2. Instantiate the BitTorrent client
	client, err := torrent.NewClient(cfg)
	if err != nil {
		<-d.maxConn
		return nil, fmt.Errorf("Error creating torrent client: %s", err)
	}
	defer client.Close()

	// 3. Add your magnet link to the client
	 // Example hash
	t, err := client.AddMagnet(d.magnetLink)
	if err != nil {
		<-d.maxConn
		return nil, err
	}

	log.Println("Connecting to peers to resolve metadata...")

	select {
		case <-t.GotInfo():
			<-d.maxConn
		case <-time.After(60 * time.Second):
			<-d.maxConn
			file.Close()
			os.Remove(file.Name())
			return nil, fmt.Errorf("timeout: could not resolve magnet link metadata within 20 seconds")
	}
	
	info := t.Metainfo()

	if err := info.Write(file); err != nil{
		file.Close()
		return nil,err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return nil, err
	}


	os.Remove(file.Name())

	return file, nil
}