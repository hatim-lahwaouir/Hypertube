package models

import (
	"time"
)

type Torrent struct {
	MovieID   uint64
	Hash      string `gorm:"primaryKey" json:"-"`
	Path      string `gorm:"not null"  json:"-"`
	Size      string `json:"size"`
	Quality   string `json:"quality"`
	Seeds     int    `gorm:"not null"  json:"seeds"`
	Peers     int    `gorm:"not null"  json:"peers"`
	CreatedAt time.Time
}

func (m *MoviesRep) CreateTorrents(t []Torrent) error {

	res := m.db.Create(t)

	return res.Error
}

func (m *MoviesRep) GetTorrent(id uint64, hash string) (*Torrent, error) {
	var (
		t Torrent
	)

	res := m.db.Model(&Torrent{}).Where("movie_id = ? and hash = ?", id, hash).First(&t)

	return &t, res.Error
}
