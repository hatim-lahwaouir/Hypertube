package handler


import (
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "fmt"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"
    "net/http"
)


type User struct  {

}


func NewUserHandler() *User {
    return &User{}
}

func (u *User) RegisterUser(w http.ResponseWriter, r *http.Request) error{
    user, field_errors := dto.NewUserSingUp(r.Body)

    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }


    fmt.Println(user)
    return utils.WriteResp(w, 200, "Account created wait for an email will be sent to you ! ")
}

func (u *User) LoginUser(){

}
