package models


import (
    "gorm.io/gorm"
    "errors"
    "time"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/dto"
    "fmt"
 )


type Movie struct {
    Id   uint64 `gorm:"primarykey"  json:"id,omitempty"`
    Name string `gorm:"unique;not null"  json:"username,omitempty"`
    IMDBCode string `gorm:"unique;not null"  json:"imdb_code,omitempty"`
    UpdatedAt time.Time
    Year int 
    Thumbnail  string
    Description string
    Rating float64 
    Torrents []Torrent `gorm:"foreignKey:MovieID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Genre []Genre `gorm:"many2many:movie_genres;"`
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






type MoviesRep struct {
    db *gorm.DB
}


func NewMoviRepository(db *gorm.DB) *MoviesRep {
    return &MoviesRep{db: db}
}





func (m *MoviesRep) CreateMovie(data *types.Movie)  error {
     
     var (
        genres []Genre
     )

     imdbGenres := types.GetGenres()


     for _, val := range data.Genres {
         genres = append(genres, Genre{Id: imdbGenres[val]})
     }

     
     result := m.db.Omit("Genre.*").Create(&Movie{Id: data.ID,
        Name: data.Title,
        IMDBCode: data.IMDBCode,
        UpdatedAt: time.Now(),
        Year: data.Year,
        Description: data.DescriptionFull,
        Rating: data.Rating,
        Genre : genres,
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
 


func FilterGenre(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
        
        if filters.Genre == "" {
			return db // Skip filtering if empty
		}
            imdbGenres  := types.GetGenres()
            genreID := imdbGenres[filters.Genre]

		return db.Joins("join movie_genres on movie_genres.movie_id = movies.id").Where("movie_genres.genre_id = ?", genreID)
    }
}


func FilterByName(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filters.Name == "" {
			return db // Skip filtering if empty
		}
		return db.Where("name LIKE ?", "%" +  filters.Name + "%")
    }
}

func OrderBy(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filters.OrderBy == "" || filters.SortBy == ""{
			return db // Skip filtering if empty
		}

        direction := "ASC"
		if filters.OrderBy == "desc" {
			direction = "DESC"
		} 
       var column string
		switch filters.SortBy {
            case "year":
                column = "year"
            case "title":
                column = "title"
            case "rating":
                column = "rating"
            default:
                return db 
		} 
		return db.Order(column + " " + direction)
    }
}

func  (m *MoviesRep) GetMoviWithGenre(filters *dto.MovieFilters) {
    var (
        movies []Movie
    )
    //res := m.db.Model(&Movie{}).Joins("join movie_genres on movie_genres.movie_id = movies.id").Where("movie_genres.genre_id = ?", genreID).Find(&movies)
    res := m.db.Model(&Movie{}).Scopes(FilterGenre(filters), FilterByName(filters), OrderBy(filters)).Find(&movies)
    fmt.Println(res.Error, movies)
}



