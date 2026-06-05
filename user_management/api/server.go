package api


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/user_management/handler"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "github.com/joho/godotenv"
    "log"
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
    // Setup Handlers 
    newUser:= handler.NewUserHandler(db)
    router := http.NewServeMux()


    router.HandleFunc("POST /SignUp", utils.MakeHandler(newUser.RegisterUser))
    router.HandleFunc("POST /ValidateEmail", utils.MakeHandler(newUser.ValidateEmail))


    server.Handler = router


    return server.ListenAndServe()
}
