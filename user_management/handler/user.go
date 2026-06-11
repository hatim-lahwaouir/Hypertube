package handler


import (
    "fmt"
    "github.com/hatim-lahwaouir/Hypertube/user_management/utils"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "github.com/hatim-lahwaouir/Hypertube/user_management/models"
    "net/http"
    "strconv"
    "time"
)


type User struct  {
    UserRep *models.UserRepository
    UserManagementService *services.UserManagementService
    AuthService *services.AuthService 
    FileUploadService *services.FileUploadService
    MaxUploadSize int64
}


func NewUserHandler(userRepo *models.UserRepository, auth *services.AuthService, fileUpload *services.FileUploadService, userManagementService *services.UserManagementService) *User {
    return &User{UserRep : userRepo, AuthService: auth, FileUploadService : fileUpload, UserManagementService : userManagementService}
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

    user, field_errors := dto.NewUserEmail(r.Body)


    if user == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }

    code, err  := u.AuthService.GenerateOneTimeCode()
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
    

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





func (u *User) GetUserInfo(w http.ResponseWriter, r *http.Request) error{

    userId := u.AuthService.GetUser(r)
    id, err := strconv.ParseUint(r.PathValue("id") ,10,64) 
    
    if err != nil {
        return utils.NewApiError(http.StatusBadRequest, "invalid id provided")
    }


    if userId.ID == id {

        userInfo, err := u.UserRep.GetCurrentUserInfo(userId)

        if err != nil {
                return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }
        return utils.WriteResp(w, http.StatusOK , userInfo)
    }
    

    userInfo,exists, err := u.UserRep.GetOtherUserInfo(dto.AuthUser{ID: id})

    if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    if exists == false {
            return utils.NewApiError(http.StatusNotFound, "Current user not found")
    }
    return utils.WriteResp(w, http.StatusOK , userInfo)
}


func (u *User) UpdateUserData(w http.ResponseWriter, r *http.Request) error{

    userId := u.AuthService.GetUser(r)
    id, err := strconv.ParseUint(r.PathValue("id") ,10,64) 
    if err != nil {
        return utils.NewApiError(http.StatusBadRequest, "invalid id provided")
    }
    if userId.ID != id {
        fmt.Println(userId.ID)
        return utils.NewApiError(http.StatusUnauthorized , "Unauthorized")
    }


    userData, field_errors := dto.NewEditUserInfo(r.Body)


    if userData == nil {
        return utils.NewApiError(http.StatusBadRequest, field_errors)
    }
    
    
    if err :=  u.UserRep.UpdateNonSensetiveData(userId, userData); err != nil {
        return utils.NewApiError(http.StatusBadRequest , "Username already exists")
    }
    // update password

        fmt.Println(userData)
    if len(userData.OldPassword)  != 0  && len(userData.NewPassword) != 0 {
        passwordHash, err := u.UserRep.GetUserPasswordWithID(userId)
        if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }
        if u.AuthService.CheckPasswordHash(userData.OldPassword , passwordHash) == false {
            return utils.NewApiError(http.StatusBadRequest, "invalid password")
        } 

        newPassword, err := u.AuthService.HashPassword(userData.NewPassword)
        if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }
       

        if  err := u.UserRep.UpdateUserPassword(userId,newPassword); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }
        // update user password

    }

    // update Email
    if len(userData.Email) != 0 {
        // check if email already exists 

        emailExists , err := u.UserRep.EmailExists(userData.Email)

        if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }

        if emailExists {
            return utils.NewApiError(http.StatusBadRequest , "Email already exists")
        }
        // generate code 
        code, err  := u.AuthService.GenerateOneTimeCode()
        if err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }
        // send to user
        if err := u.UserManagementService.SendChangeEmail(userData.Email, code); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }


        // save code to database
        if  err := u.UserRep.UpsertUserChangeEmailCode(userId,userData.Email, code); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
        }

    }
    
     // install an image if it was provided 
    if len(userData.ProfilePic) != 0 {

            fmt.Println(userData.ProfilePic)
            pic_name , err := u.FileUploadService.DownloadAnImage(userData.ProfilePic)
            if err != nil {
                return utils.NewApiError(http.StatusBadRequest , err.Error())
            }
            if err := u.UserRep.UpdateUserPic(userId, pic_name); err != nil {
                return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
            }
    }

    return utils.WriteResp(w, http.StatusOK , "ok")
}


func (u *User) ValidateMyNewEmail(w http.ResponseWriter, r *http.Request) error{
    var (
        duration time.Duration
        code *dto.UserCode 
    )
    userId := u.AuthService.GetUser(r)
    code = dto.NewUserCode(r.Body)

    duration , _ = time.ParseDuration("24h00m0s")
    if code == nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

     user_model, exists, err := u.UserRep.GetUserCodeChangeEmail(userId)

     if err != nil  {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
     }
     if exists == false  {
            return utils.NewApiError(http.StatusBadRequest , "there is no change email")
     }

     if code.Code != user_model.Code {
            return utils.NewApiError(http.StatusBadRequest , "Invalid Code")
     }

    // delete record
    if err := u.UserRep.DeleteUserChangeEmail(userId); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    if user_model.UpdatedAt.Sub(time.Now()).Abs() >  duration {
            return utils.NewApiError(http.StatusUnauthorized , "your code has expired")
    }

    // change User Email

    if err := u.UserRep.ChangeUserEmail(userId,user_model.NewEmail); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }
   
    return utils.WriteResp(w, http.StatusOK , "email changed")
}


func (u *User) UploadPic(w http.ResponseWriter, r *http.Request) error{
    userId := u.AuthService.GetUser(r)

    r.ParseMultipartForm(u.MaxUploadSize)
    pic , _ , err := r.FormFile("profile_pic")
    if err != nil {
        return utils.NewApiError(http.StatusBadRequest, "Error retrieving profile picture")
    }
    defer pic.Close()

    pic_name ,  err := u.FileUploadService.UploadFile(pic)

    if err != nil {
        return utils.NewApiError(http.StatusBadRequest, err.Error())
    }

    if err := u.UserRep.UpdateUserPic(userId, pic_name); err != nil {
            return utils.NewApiError(http.StatusInternalServerError, "Internal Server Error")
    }

    return utils.WriteResp(w, http.StatusCreated, "imag uploaded !")
}


