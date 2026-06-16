package models




import (

  "gorm.io/driver/postgres"
  "gorm.io/gorm/logger"
  "time"

 "gorm.io/gorm"
  "fmt"
  "os"
  "log"
)







func StartDb() *gorm.DB {


	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // IO writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Info,   // Log level set to Info for all queries
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,         // Don't include params in the SQL log if true
			Colorful:                  true,          // Enable color printing
		},
	)

    dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", 
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"), 
        os.Getenv("PORT"))
       

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger,})


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
