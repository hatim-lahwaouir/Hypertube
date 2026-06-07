package models




import (
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "fmt"
  "os"
  "log"
)


func StartDb() *gorm.DB {



    dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", 
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"), 
        os.Getenv("PORT"))
       

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})


    // migration
    db.AutoMigrate(&UserEmail{})
    db.AutoMigrate(&User{})

    if err != nil {
        log.Fatal("failed to connect database", err)
    }

    return db
}
