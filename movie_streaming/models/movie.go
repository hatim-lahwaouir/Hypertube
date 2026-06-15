package models


import (
    "gorm.io/gorm"
    "errors"
    "time"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
 )


type Movie struct {
    Id   uint64 `gorm:"primarykey"  json:"id,omitempty"`
    Name string `gorm:"unique;not null"  json:"username,omitempty"`
    IMDBCode string `gorm:"unique;not null"  json:"imdb_code,omitempty"`
    //TorrentPath string `gorm:"unique;not null"  json:"-"`
    UpdatedAt time.Time
    Year int 
    Thumbnail  string
    Description string
    Rating float64 
    Torrents []Torrent `gorm:"foreignKey:MovieID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Genre []Genre `gorm:"many2many:movie_genre;"`
}



type Torrent struct {
    MovieID uint64
    Path    string `gorm:"not null"  json:"-"`
    Size   string `json:"size"`
    Quality string `json:"quality"`
    Seeds int `gorm:"not null"  json:"seeds"`
    Peers int `gorm:"not null"  json:"peers"`
    CreatedAt time.Time
}


type Genre struct {
    gorm.Model
    Type string `gorm:"unique;not null" json:"type"`
}



type MoviesRep struct {
    db *gorm.DB
}


func NewMoviRepository(db *gorm.DB) *MoviesRep {
    return &MoviesRep{db: db}
}





func (m *MoviesRep) CreateMovie(data *types.Movie)  error {
     

     result := m.db.Create(&Movie{Id: data.ID,
        Name: data.Title,
        IMDBCode: data.IMDBCode,
        UpdatedAt: time.Now(),
        Year: data.Year,
        Description: data.DescriptionFull,
        Rating: data.Rating,
        Thumbnail: data.LargeCoverImage,

     })
     return result.Error
}

 
func (m *MoviesRep) MovieExists(id uint64)  (bool, error) {
    var (
        exists int64
    )

    exists = 0
    res := m.db.Model(&Movie{}).Where("id = ?", id).Count(&exists)

    return exists == 1, res.Error
}



func (m *MoviesRep) GetMovie(id uint64)  (*Movie,bool, error) {
    var (
        movie Movie 
    )

    res := m.db.Model(&Movie{}).Where("id = ?", id).First(&movie)

    if res.Error != nil {
        if errors.Is(res.Error, gorm.ErrRecordNotFound) {
            return nil, false,  nil
        } else {
             return nil, false,  res.Error
        }
    }

    return  &movie ,true, nil 
}
     
func (m *MoviesRep) CreateTorrents(t []Torrent)  error {

    res := m.db.Create(t)

    return res.Error 
}
 



