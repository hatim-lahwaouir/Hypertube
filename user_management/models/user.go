package models


import (
    "time"
    "errors"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
 )




type UserEmail struct {
    UpdatedAt time.Time 
    Code  string
    Email string `gorm:"unique;not null"`
}


type User struct {
    ID           uint64 `gorm:"primarykey"`
    Username string  `gorm:"unique;not null"`
    Email string `gorm:"unique;not null"`
    Password string `gorm:"not null"` 
    FirstName string  `gorm:"not null"`
    LastName string  `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null"`
}


type UserRepository struct {
    db *gorm.DB
}


func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}


func (u * UserRepository) UserNameExists(username string)  (bool, error) {
    var (
        exists int64
    )

    exists = 0
    res := u.db.Model(&User{}).Where("username = ?", username).Count(&exists)

    return exists == 1, res.Error
}

func (u * UserRepository) EmailExists(email string)  (bool, error) {
    var (
        exists int64
    )

    exists = 0

    res := u.db.Model(&User{}).Where("email = ?", email).Count(&exists)

    return exists == 1, res.Error
}



func (u * UserRepository) CreateUser(user dto.UserSignUp)  error {
        result := u.db.Create(&User{Password: user.Password, 
        Username: user.Username,
       Email: user.Email, 
        FirstName : user.FirstName,
        LastName: user.LastName})

        return result.Error
}



func (u * UserRepository) GetUserCodeEmail(email string)  (UserEmail,bool, error) {
    var (
        user_code UserEmail
    )
    res  := u.db.First(&user_code, "email = ?", email)
   
   if errors.Is(res.Error, gorm.ErrRecordNotFound) {
       return user_code, false, nil
   }


    return user_code,true, res.Error 
}

func (u * UserRepository) DeleteUserCodeEmail(email string)  error {
    

    res := u.db.Where("email = ?", email).Delete(&UserEmail{})

    return res.Error 
}


func (u *UserRepository) UpsertUserEmailCode(user dto.UserEmail, code string) error {


        res := u.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "email"}},
        DoUpdates: clause.AssignmentColumns([]string{"code", "updated_at"}),
        }).Create(&UserEmail{Code: code, Email: user.Email, UpdatedAt: time.Now()})


        return res.Error 
}


func (u *UserRepository) GetUserPassword(user dto.UserLogin)  (User, bool , error) {
    var (
        user_model User
    )

    res := u.db.Model(&User{}).Select("id", "email", "password").Where("email = ?", user.Email).First(&user_model)

    if res.Error != nil {
        if errors.Is(res.Error, gorm.ErrRecordNotFound) {
            return user_model, false, nil
        } else {
             return user_model, false, res.Error

        }
    }


    return  user_model,true , nil 
}



func (u *UserRepository) GetCurrentUserInfo (user dto.AuthUser)  (*dto.CurrentUserInfo , error) {
    var (
        user_model User
    )

    res := u.db.Model(&User{}).Select("id", "email", "username", "first_name", "last_name").Where("id = ?", user.ID).First(&user_model)

    if res.Error != nil {
        return nil,  res.Error
    }


    return  &dto.CurrentUserInfo{ID: user_model.ID, Username: user_model.Username, LastName: user_model.LastName, FirstName: user_model.LastName, Email: user_model.Email} , nil
}
