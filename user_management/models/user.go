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
    ID           uint64 `gorm:"primarykey"  json:"id,omitempty"`
    Username string  `gorm:"unique;not null"  json:"username,omitempty"`
    Email string `gorm:"unique;not null"  json:"email,omitempty"`
    Password string `gorm:"not null" json:"-"` 
    FirstName string  `gorm:"not null" json:"first_name,omitempty"`
    LastName string  `gorm:"not null" json:"last_name,omitempty"`
    ProfilePic string `gorm:"default:default.png" json:"profile_pic,omitempty"`
    CreatedAt time.Time `gorm:"not null" json:"-"`
    NewEmail UserChangeEmail  ` gorm:"foreignKey:UserID" json:"new_email,omitempty" `
}

type UserChangeEmail struct {
    UpdatedAt time.Time  `json:"-"` 
    Code  string `json:"-"` 
    NewEmail string `gorm:"unique;not null" json:"email,omitempty"` 
    UserID uint64  `gorm:"unique;not null" json:"-"`  
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


func (u *UserRepository) UpsertUserChangeEmailCode( user dto.AuthUser, email string, code string) error {
        
        res := u.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "user_id"}},
        DoUpdates: clause.AssignmentColumns([]string{"new_email","code", "updated_at"}),
        }).Create(&UserChangeEmail{UserID: user.ID, Code: code, NewEmail: email, UpdatedAt: time.Now()})


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



func (u *UserRepository) GetCurrentUserInfo (user dto.AuthUser)  (*User , error) {
    var (
        user_model User
    )

    res := u.db.Model(&User{}).Select("id", "email", "username", "first_name", "last_name", "profile_pic").Where("id = ?", user.ID).Preload("NewEmail").First(&user_model)

    if res.Error != nil {
        return nil,  res.Error
    }


    return  &user_model , nil
}

func (u *UserRepository) GetOtherUserInfo (user dto.AuthUser)  (*User ,bool,  error) {
    var (
        user_model User
    )
    
    res := u.db.Model(&User{}).Select("id", "username", "profile_pic").Where("id = ?", user.ID).First(&user_model)
    if errors.Is(res.Error, gorm.ErrRecordNotFound) {
       return nil, false, nil
    }
    if res.Error != nil {
        return nil,false,  res.Error
    }


    return &user_model, true, nil 
}

func (u *UserRepository) UpdateUserPic(user dto.AuthUser, profile_pic_path string)  error {
    

    res := u.db.Model(&User{}).Where("id = ?", user.ID ).Update("profile_pic",profile_pic_path)

    return res.Error
}



func (u *UserRepository) UpdateNonSensetiveData(user dto.AuthUser, userData *dto.EditUserInfo)   error {

    res := u.db.Model(&User{ID: user.ID}).Updates(User{Username : userData.Username, ProfilePic : userData.ProfilePic})


    return  res.Error 
}


func (u *UserRepository) UpdateUserPassword(user dto.AuthUser, hashedPassword string)   error {

    res := u.db.Model(&User{ID: user.ID}).Updates(User{Password : hashedPassword})

    return  res.Error 
}

func (u *UserRepository) GetUserPasswordWithID(user dto.AuthUser)  (string , error) {
    var (
        user_model User
    )

    res := u.db.Model(&User{}).Select("id", "password").Where("id = ?", user.ID).First(&user_model)

    if res.Error != nil {
        if errors.Is(res.Error, gorm.ErrRecordNotFound) {
            return "", nil
        } else {
             return "", res.Error

        }
    }


    return  user_model.Password, nil 
}


func (u *UserRepository) GetUserCodeChangeEmail(user dto.AuthUser)  (*UserChangeEmail , bool ,error) {
    var (
        user_model UserChangeEmail
    )

    res := u.db.Model(&UserChangeEmail{}).Where("user_id = ?", user.ID).First(&user_model)

    if res.Error != nil {
        if errors.Is(res.Error, gorm.ErrRecordNotFound) {
            return nil, false,  nil
        } else {
             return nil, false,  res.Error

        }
    }

    return  &user_model,true, nil 
}

func (u *UserRepository) ChangeUserEmail(user dto.AuthUser, email string)  (error) {
    
    res := u.db.Model(&User{ID: user.ID}).Updates(User{Email : email})

    return  res.Error 
}


func (u *UserRepository) DeleteUserChangeEmail(user dto.AuthUser)  (error) {
    res := u.db.Where("user_id = ?", user.ID).Delete(&UserChangeEmail{})

    return res.Error
}



