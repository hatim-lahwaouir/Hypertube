package handler


import (
    "fmt"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "net/http"
    "time"
)


type User struct  {
    UserRep *models.UserRepository
    AuthService *services.AuthService 
}


func NewUserHandler(userRepo *models.UserRepository, auth *services.AuthService ) *User {
    return &User{UserRep : userRepo, AuthService: auth}
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

    // check if username already taken
    //u.db.Model(&User{}).Where("username = ?", user.Username).Count(&exists)
    
    UsernameExists, err := u.UserRep.UserNameExists(user.Username) 
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    // check if email already exists 
    EmailExists, err := u.UserRep.EmailExists(user.Email)
    if err !=  nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    } 

    if EmailExists && UsernameExists {
        return utils.NewApiError(http.StatusBadRequest , "Email And Username  already registred")
    }


    if UsernameExists {
        return utils.NewApiError(http.StatusBadRequest , "Username already registred")
    }
    if EmailExists {
        return utils.NewApiError(http.StatusBadRequest , "Email already registred")
    }

    // get record first 
    user_code,userExists,  err := u.UserRep.GetUserCodeEmail(user.Email)

    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    
    if userExists == false {
        return utils.NewApiError(http.StatusBadRequest , "this email doesn't exists, resend code to this email")
    }

    // delete record
    //u.db.Where("email = ?", user.Email).Delete(&models.UserEmail{})
    if err = u.UserRep.DeleteUserCodeEmail(user.Email); err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    

    if user_code.UpdatedAt.Sub(time.Now()).Abs() >  duration {
            return utils.NewApiError(http.StatusUnauthorized , "your code has expired")
    }
    if user_code.Code != user.Code {
        return utils.NewApiError(http.StatusUnauthorized , "Invalid code")
    }


    

    hashedPassword , err := u.AuthService.HashPassword(user.Password)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    user.Password = hashedPassword
    
    err = u.UserRep.CreateUser(*user)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

        

    return utils.WriteResp(w, http.StatusCreated, "Account created wait for an email will be sent to you ! ")
}


func (u *User) ValidateEmail(w http.ResponseWriter, r *http.Request) error{
    var (
        code string 

    )

    user, field_errors := dto.NewUserEamil(r.Body)


    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }

    code = u.AuthService.GenerateOneTimeCode()
    

    if err := u.UserRep.UpsertUserEmailCode(*user, code); err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    if u.AuthService.SendCodeViaMail(user.Email, code) != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    return utils.WriteResp(w, http.StatusCreated, "we sent you a code to validate your email !")
}



func (u *User) Login(w http.ResponseWriter, r *http.Request) error{
    user, field_errors := dto.NewUserLogin(r.Body)
    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }

    //res := u.db.Model(&models.User{}).Select("id", "email", "password").Where("email = ?", user.Email).First(&user_model)

    user_model,userExists , err := u.UserRep.GetUserPassword(*user)
    if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    if userExists == false {
         return utils.NewApiError(http.StatusUnauthorized, "Invalid credentials")
    }

    

    fmt.Println(user_model.ID)
    if u.AuthService.CheckPasswordHash(user.Password, user_model.Password) ==  false {
        return utils.NewApiError(http.StatusUnauthorized, "Invalid credentials")
    }
    
    token , err  := u.AuthService.GenerateJWT(user_model.ID)
    if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    return utils.WriteResp(w, http.StatusCreated, dto.NewJwt(token))
}





func (u *User) GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) error{
    userId := u.AuthService.GetUser(r)

    userInfo, err := u.UserRep.GetCurrentUserInfo(userId)

    if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    return utils.WriteResp(w, http.StatusOK , userInfo)
}



