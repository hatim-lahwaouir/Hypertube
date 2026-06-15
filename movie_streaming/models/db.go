package models




import (

  "gorm.io/gorm/clause"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "fmt"
  "os"
  "log"
)



var imdbGenres []string = []string{ 
  "Action",
  "Adult",
  "Adventure",
  "Animation",
  "Biography",
  "Comedy",
  "Crime",
  "Documentary",
  "Drama",
  "Family",
  "Fantasy",
  "Film-Noir",
  "Game-Show",
  "History",
  "Horror",
  "Music",
  "Musical",
  "Mystery",
  "News",
  "Reality-TV",
  "Romance",
  "Sci-Fi",
  "Short",
  "Sport",
  "Talk-Show",
  "Thriller",
  "War",
  "Western",
}


func setupGenres(db *gorm.DB)  error {
    var (
        genre_model []Genre
    )



    for _, val := range(imdbGenres){
        genre_model = append(genre_model, Genre{Type: val})
    }

	res := db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&genre_model)


    return res.Error 
}

func StartDb() *gorm.DB {



    dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", 
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"), 
        os.Getenv("PORT"))
       

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})


    //migration
    db.AutoMigrate(&Movie{})
    db.AutoMigrate(&Genre{})
    db.AutoMigrate(&Torrent{})


    // setup generes 
    setupGenres(db)


    if err != nil {
        log.Fatal("failed to connect database", err)
    }

    return db
}
