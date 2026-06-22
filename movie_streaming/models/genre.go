package models

import (

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/types"
)

type Genre struct {
    Id        uint64           `gorm:"primaryKey"`
    Type string                `gorm:"unique;not null" json:"type"`
}

func setupGenres(db *gorm.DB)  error {
    var (
        genre_model []Genre
    )

    imdbGenres := types.GetGenres()



    for val, id:= range(imdbGenres){
        genre_model = append(genre_model, Genre{Id:uint64(id), Type: val})
    }

	res := db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&genre_model)


    return res.Error 
}



