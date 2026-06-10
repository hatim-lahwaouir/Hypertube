package api


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/user_management/middleware"
    "github.com/hatim-lahwaouir/Hypertube/user_management/handler"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "fmt"
    "os"
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
    maxUploadSize := int64(5 * 1024 * 1024)
    // setup services 
    authService := services.NewAuthService()
    fileUploadService := services.NewFileUploadService(os.Getenv("FILE_UPLOAD_PATH"), maxUploadSize) 
    // setup respositories 
    userRepository := models.NewUserRepository(db)


    // Setup Handlers 

    newUser:= handler.NewUserHandler(userRepository, authService, fileUploadService)
    router := http.NewServeMux()
    authRouter  :=  http.NewServeMux()


    router.HandleFunc("POST /signUp", utils.MakeHandler(newUser.RegisterUser))
    router.HandleFunc("POST /validateEmail", utils.MakeHandler(newUser.ValidateEmail))
    router.HandleFunc("POST /login", utils.MakeHandler(newUser.Login))


    // routes that need authentication 
    authRouter.HandleFunc("GET /users/{id}", utils.MakeHandler(newUser.GetUserInfo))
    authRouter.HandleFunc("PATCH /users/{id}", utils.MakeHandler(newUser.UpdateUserData))
    authRouter.HandleFunc("POST /upload", utils.MakeHandler(newUser.UploadPic))

    router.Handle("/", middleware.Auth(authRouter))
    server.Handler = router


    return server.ListenAndServe()
}
