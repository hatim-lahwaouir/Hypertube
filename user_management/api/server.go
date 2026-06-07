package api


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/user_management/middleware"
    "github.com/hatim-lahwaouir/Hypertube/user_management/handler"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "fmt"
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
    fmt.Println("start db")
    db := models.StartDb() 
    // setup services 
    authService := services.NewAuthService()
    // setup respositories 
    userRepository := models.NewUserRepository(db)


    // Setup Handlers 

    newUser:= handler.NewUserHandler(userRepository, authService)
    router := http.NewServeMux()
    authRouter  :=  http.NewServeMux()


    router.HandleFunc("POST /signUp", utils.MakeHandler(newUser.RegisterUser))
    router.HandleFunc("POST /validateEmail", utils.MakeHandler(newUser.ValidateEmail))
    router.HandleFunc("POST /login", utils.MakeHandler(newUser.Login))


    // routes that need authentication 
    authRouter.HandleFunc("GET /me", utils.MakeHandler(newUser.GetCurrentUserInfo))

    router.Handle("/", middleware.Auth(authRouter))
    server.Handler = router


    return server.ListenAndServe()
}
