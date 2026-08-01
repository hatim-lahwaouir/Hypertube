package services

import (
	"io"
	"log"
	"os"

	"github.com/anacrolix/torrent"
)





type  DownloadTorrent struct {
	magnetLink string
}



func NewDownloadTorrent(m string) *DownloadTorrent{
	return &DownloadTorrent{magnetLink:  m}
}





func (d *DownloadTorrent) DownloadTorrent() (*os.File, error) {

	file, err := os.CreateTemp("","file.*.torrent")
	if err != nil {
		return nil, err
	}


	cfg := torrent.NewDefaultClientConfig()
	
	// Optional: Set a dummy data storage directory so it doesn't download actual media files
	cfg.DataDir = os.TempDir()

	// 2. Instantiate the BitTorrent client
	client, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating torrent client: %s", err)
	}
	defer client.Close()

	// 3. Add your magnet link to the client
	 // Example hash
	t, err := client.AddMagnet(d.magnetLink)
	if err != nil {
		log.Fatalf("Error adding magnet link: %s", err)
	}

	log.Println("Connecting to peers to resolve metadata...")

	<-t.GotInfo()

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