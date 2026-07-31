package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/handler"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/services"
	"github.com/hatim-lahwaouir/Hypertube/movie_streaming/utils"
	"github.com/joho/godotenv"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow all origins for development
        w.Header().Set("Access-Control-Allow-Origin", "*")
        
        // Allow specific methods
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        
        // IMPORTANT: "Range" must be allowed for video streaming to work via CORS!
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, Range")

        // Intercept preflight OPTIONS requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Move to the next handler
        next.ServeHTTP(w, r)
    })
}


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

	StreamingService := services.NewMovieStreamingService()
	// handlers
	//movieHandler := handler.NewMovieHandler(downalodService)
	movieStreamHandler := handler.NewMovieStreamingHandler(StreamingService)

	router := http.NewServeMux()

	router.HandleFunc("POST /movie/", utils.MakeHandler(movieStreamHandler.Download))
	router.HandleFunc("GET /movie/{infohash}", utils.MakeHandler(movieStreamHandler.StreamVideo))

	server.Handler = corsMiddleware(router)
	fmt.Println("here")
	return server.ListenAndServe()
}
