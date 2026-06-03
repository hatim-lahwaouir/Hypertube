package api


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/user_management/handler"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
)






func NewServer(lAddr string) http.Server{
    return http.Server {
        Addr: lAddr,
    }
}



func StartServer(server http.Server) error {
    // Setup Handlers 
    newUser:= handler.NewUserHandler()
    router := http.NewServeMux()


    router.HandleFunc("POST /hello", utils.MakeHandler(newUser.RegisterUser))


    server.Handler = router


    return server.ListenAndServe()
}
