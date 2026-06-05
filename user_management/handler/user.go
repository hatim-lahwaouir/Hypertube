package handler


import (
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "fmt"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "net/http"
    "gorm.io/gorm"
    "time"
)


type User struct  {
    db *gorm.DB

}


func NewUserHandler(db *gorm.DB) *User {
    return &User{db : db }
}

func (u *User) RegisterUser(w http.ResponseWriter, r *http.Request) error{
    var (
        user_code models.UserEmail
        duration time.Duration

    )
    user, field_errors := dto.NewUserSingUp(r.Body)
    duration , _ = time.ParseDuration("0h10m0s")

    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }

    u.db.First(&user_code, "email = ?", user.Email)

    fmt.Println(user_code.CreatedAt.Sub(time.Now()).Abs(), duration)


    if user_code.CreatedAt.Sub(time.Now()).Abs() >  duration {
        return utils.NewApiError(http.StatusUnauthorized , "your code has expired")
    }

    if user_code.Code != user.Code {
        return utils.NewApiError(http.StatusUnauthorized , "Invalid code")
    }


    fmt.Println(user_code)

    return utils.WriteResp(w, 200, "Account created wait for an email will be sent to you ! ")
}


func (u *User) ValidateEmail(w http.ResponseWriter, r *http.Request) error{
    var (
        code string 

    )

    user, field_errors := dto.NewUserEamil(r.Body)


    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }

    code = services.GenerateOneTimeCode()
    


    u.db.Create(&models.UserEmail{Code: code, Email: user.Email})

    if services.SendCodeViaMail(user.Email, code) != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    return utils.WriteResp(w, 200, "we sent you a code to validate your email !")
}
