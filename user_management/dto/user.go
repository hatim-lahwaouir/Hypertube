package dto


import (
    "io"
    "encoding/json"
    "errors"
    "github.com/go-playground/validator/v10"
    "strings"
)




type UserSignUp struct{
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"` 
    Email string   `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=10,max=50,strong_password"` 
    FirstName string `json:"firstname" validate:"required,min=3,max=50,alpha"`
    LastName string `json:"lastname" validate:"required,min=3,max=50,alpha"`
    Code     string  `json:"code" validate:"required"`
}

type CurrentUserInfo struct{
    ID  uint64
    Username string `json:"username,omitempty" validate:"required,min=3,max=50,alphanum"` 
    Email string   `json:"email,omitempty" validate:"required,email"`
    FirstName string `json:"firstname,omitempty" validate:"required,min=3,max=50,alpha"`
    LastName string `json:"lastname,omitempty" validate:"required,min=3,max=50,alpha"`
    ProfilePic  string   `json:"profile_pic,omitempty" validate:"required,email"`
}


type  AuthUser struct {
    ID  uint64
}

type  UserCode struct {
    Code     string  `json:"code" validate:"required"`
}

type UserEmail struct{
    Email string   `json:"email" validate:"required,email"`
}

type UserLogin struct{
    Email    string   `json:"email" validate:"required,email"`
    Password string   `json:"password" validate:"required"`
}

type EditUserInfo struct{
    Email        string   `json:"email" validate:"omitempty,email"`
    Username     string   `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
    ProfilePic   string   `json:"profile_pic" validate:"omitempty,url"`
    OldPassword  string   `json:"old_password" validate:"required_with=NewPassword,omitempty,min=10,max=50,strong_password"`
    NewPassword  string   `json:"password" validate:"omitempty,min=10,max=50,strong_password"`
}

var Errors map[string]string  = map[string]string {
    "Username" : "required, min len 3, max len 50 only conaitns alphanumeric",
    "Email" :  "Invalid email",
    "Password" : "required, min len 10, max len 50, must conatins alphanumeric, special characters, lower case letter and upper case letters  ",
    "OldPassword" : "required with new password, min len 10, max len 50, must conatins alphanumeric, special characters, lower case letter and upper case letters  ",
    "FirstName" : "required, only alpha , min len 3 and max len 50",
    "LastName" : "required, only alpha , min len 3 and  max len 50",
}




func NewUserCode(body io.Reader) (*UserCode) {
    var (
        user UserCode
    )


    json.NewDecoder(body).Decode(&user)
    err := Validate.Struct(user)
    if err != nil {
        return nil
    }
    return &user
}


func NewUserSingUp(body io.Reader) (*UserSignUp, map[string]string) {
    
    var (
        user UserSignUp
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&user)
    


    err := Validate.Struct(user)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }

    return &user, nil
}

func NewEditUserInfo(body io.Reader) (*EditUserInfo, map[string]string) {
    
    var (
        user EditUserInfo
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&user)
    


    err := Validate.Struct(user)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }

    return &user, nil
}

func NewUserLogin (body io.Reader) (*UserLogin, map[string]string) {
    
    var (
        user UserLogin
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&user)
    


    err := Validate.Struct(user)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }

    return &user, nil
}


func NewUserEmail(body io.Reader) (*UserEmail, map[string]string) {
    
    var (
        user UserEmail
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&user)
    


    err := Validate.Struct(user)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }
    return &user, nil
}


