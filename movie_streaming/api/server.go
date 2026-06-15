package api


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/handler"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
    "github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
    "log"
    "github.com/joho/godotenv"
)






func NewServer(lAddr string) http.Server{
    return http.Server {
        Addr: lAddr,
    }
}



func StartServer(server http.Server) error {
    // load .env
   
   	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
    // start db
    db := models.StartDb() 

    // models repositories

    movieRep := models.NewMoviRepository(db)
    //services 

    downalodService := services.NewDownloadMovieService(movieRep)
    // handlers
    movieHandler := handler.NewMovieHandler(downalodService)


    router := http.NewServeMux()


    log.Println("db up",db)
    router.HandleFunc("POST /Hello", utils.MakeHandler(movieHandler.Hello))
    router.HandleFunc("POST /FilterMovie", utils.MakeHandler(movieHandler.MovieSuggestions))

    server.Handler = router


    return server.ListenAndServe()
}
