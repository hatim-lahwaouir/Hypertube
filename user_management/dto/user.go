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

type UserEmail struct{
    Email string   `json:"email" validate:"required,email"`
}

var Errors map[string]string  = map[string]string {
    "Username" : "required, min len 3, max len 50 only conaitns alphanumeric",
    "Email" :  "Invalid email",
    "Password" : "required, min len 10, max len 50, must conatins alphanumeric, special characters, lower case letter and upper case letters  ",
    "FirstName" : "required, only alpha , min len 3 and max len 50",
    "LastName" : "required, only alpha , min len 3 and  max len 50",
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



func NewUserEamil(body io.Reader) (*UserEmail, map[string]string) {
    
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


