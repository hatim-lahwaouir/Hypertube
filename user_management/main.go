package main 

import (
    "github.com/hatim-lahwaouir/Hypertube/user_management/api"
    "fmt"
)



func main(){
       s := api.NewServer(":8080")
       fmt.Println("Server listening on port 8080")
       api.StartServer(s)
}
