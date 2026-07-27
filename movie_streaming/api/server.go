package api

import (
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/handler"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/models"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func NewServer(lAddr string) http.Server {
	return http.Server{
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

	downalodService := services.NewDownloadMovieInfoService(movieRep)
	StreamingService := services.NewMovieStreamingService()
	// handlers
	movieHandler := handler.NewMovieHandler(downalodService)
	movieStreamHandler := handler.NewMovieStreamingHandler(StreamingService)

	router := http.NewServeMux()

	log.Println("db up", db)
	router.HandleFunc("GET /movie/{imdb_code}", utils.MakeHandler(movieHandler.Hello))
	router.HandleFunc("POST /search-movies/{page}", utils.MakeHandler(movieHandler.MovieSuggestions))
	router.HandleFunc("POST /search-movies-omdb/{page}", utils.MakeHandler(movieHandler.MovieSuggersionsOMDB))

	router.HandleFunc("POST /movie/{id}/{hash}", utils.MakeHandler(movieStreamHandler.Download))

	server.Handler = router

	return server.ListenAndServe()
}
