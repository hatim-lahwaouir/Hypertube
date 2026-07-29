package models

import (
	"errors"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/dto"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
	"gorm.io/gorm"
	"time"
)

type Movie struct {
	Id          uint64 `gorm:"primarykey"  json:"id,omitempty"`
	Name        string `gorm:"unique;not null"  json:"name,omitempty"`
	IMDBCode    string `gorm:"unique;not null"  json:"imdb_code,omitempty"`
	UpdatedAt   time.Time
	Year        int
	Thumbnail   string
	Description string
	Rating      float64
	Torrents    []Torrent `gorm:"foreignKey:MovieID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Genre       []Genre   `gorm:"many2many:movie_genres;"`
}

type MoviesRep struct {
	db *gorm.DB
}

func NewMoviRepository(db *gorm.DB) *MoviesRep {
	return &MoviesRep{db: db}
}

func (m *MoviesRep) CreateMovie(data *types.Movie) error {

	var (
		genres []Genre
	)

	imdbGenres := types.GetGenres()

	for _, val := range data.Genres {
		genres = append(genres, Genre{Id: imdbGenres[val]})
	}

	result := m.db.Omit("Genre.*").Create(&Movie{Id: data.ID,
		Name:        data.Title,
		IMDBCode:    data.IMDBCode,
		UpdatedAt:   time.Now(),
		Year:        data.Year,
		Description: data.DescriptionFull,
		Rating:      data.Rating,
		Genre:       genres,
		Thumbnail:   data.LargeCoverImage,
	})
	return result.Error
}

func (m *MoviesRep) CreateMovies(movies []types.Movie) error {

	var (
		movies_model []Movie
	)

	imdbGenres := types.GetGenres()

	for _, data := range movies {

		var genres []Genre
		for _, val := range data.Genres {
			genres = append(genres, Genre{Id: imdbGenres[val]})
		}
		m := Movie{Id: data.ID,
			Name:        data.Title,
			IMDBCode:    data.IMDBCode,
			UpdatedAt:   time.Now(),
			Year:        data.Year,
			Description: data.DescriptionFull,
			Rating:      data.Rating,
			Genre:       genres,
			Thumbnail:   data.LargeCoverImage,
		}
		movies_model = append(movies_model, m)
	}
	result := m.db.Omit("Genre.*").Create(movies_model)
	return result.Error
}

func (m *MoviesRep) GetMovie(imdb_code string) (*Movie, bool, error) {
	var (
		movie Movie
	)

	res := m.db.Preload("Torrents").Where("movies.imdb_code = ?", imdb_code).First(&movie)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, false, nil
		} else {
			return nil, false, res.Error
		}
	}


	return &movie, true, nil
}

func FilterGenre(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		if filters.Genre == "" {
			return db // Skip filtering if empty
		}
		imdbGenres := types.GetGenres()
		genreID := imdbGenres[filters.Genre]

		return db.Joins("join movie_genres on movie_genres.movie_id = movies.id").Where("movie_genres.genre_id = ?", genreID)
	}
}

func FilterByName(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filters.Name == "" {
			return db // Skip filtering if empty
		}
		return db.Where("name LIKE ?", "%"+filters.Name+"%")
	}
}

func OrderBy(filters *dto.MovieFilters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filters.OrderBy == "" {
			return db
		}

		direction := "ASC"
		if filters.OrderBy == "desc" {
			direction = "DESC"
		}
		var column string
		switch filters.SortBy {
		case "year":
			column = "year"
		case "name":
			column = "name"
		default:
			column = "rating"
		}
		return db.Order(column + " " + direction)
	}
}

func (m *MoviesRep) GetMoviesWithFilters(filters *dto.MovieFilters, pagination uint64) ([]Movie, error) {
	var (
		movies []Movie
	)
	if pagination == 0 {
		pagination = 1
	}
	pageSize := 10

	offset := (int(pagination) - 1) * pageSize

	res := m.db.Model(&Movie{}).Scopes(FilterGenre(filters), FilterByName(filters), OrderBy(filters)).Joins("INNER JOIN torrents ON torrents.movie_id = movies.id").Preload("Genre").Offset(offset).Limit(pageSize).Find(&movies)

	return movies, res.Error
}
