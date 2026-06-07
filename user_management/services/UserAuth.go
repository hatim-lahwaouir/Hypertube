package services

import (
    gomail "gopkg.in/mail.v2"
    "os"
    "github.com/golang-jwt/jwt/v5"
    "github.com/hatim-lahwaouir/Hypertube/user_management/dto"
    "time"
    "math/rand"
    "strconv"
    "net/http"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "errors"
)



type AuthService  struct {
    ownerEmail string
    sendEmailPassword string
    jwtSecret []byte 

}


var cacheAuthService *AuthService


func NewAuthService () *AuthService {

    if cacheAuthService == nil {

        cacheAuthService =  &AuthService{ownerEmail: os.Getenv("OWNER"), sendEmailPassword: os.Getenv("EMAIL_PASSWORD"), jwtSecret: []byte(os.Getenv("JWT_SECRET")) }
    } 


    return  cacheAuthService



}


var (
    ErrInvalidToken       = errors.New("invalid token")
    ErrExpiredToken       = errors.New("token has expired")
)






func (auth *AuthService) SendCodeViaMail(email string, code string) error {

    message := gomail.NewMessage()

    // Set email headers
    message.SetHeader("From", auth.ownerEmail)
    message.SetHeader("To", email)
    message.SetHeader("Subject", "This is an email sent via Gomail and Gmail SMTP")

    // Set email body
    message.SetBody("text/plain", fmt.Sprintf("code for email validation '%s'", code))

    // Set up the SMTP dialer
    dialer := gomail.NewDialer("smtp.gmail.com", 587, auth.ownerEmail, auth.sendEmailPassword)

    // Send the email
    if err := dialer.DialAndSend(message); err != nil {
        fmt.Println("Error:", err)
        return err 
    } 
    return nil
}




func (auth *AuthService) GenerateOneTimeCode() string {
    var (
        nbr int
    )


    rand.Seed(time.Now().UnixNano())
    min := 999999 
    max := 9999999 
    nbr = rand.Intn(max - min + 1) + min
    return strconv.Itoa(nbr)
}


func (auth *AuthService) GenerateJWT(userID uint64) (string, error) {
    expirationTime := time.Now().Add(time.Hour * 24)

    claims := jwt.MapClaims{
        "sub":      strconv.FormatUint(userID, 10),      
        "exp":      expirationTime.Unix(), // expiration time
        "iat":      time.Now().Unix(),     // issued at time
    }

    // Create the token with claims
    token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims )
    tokenString, err := token.SignedString(auth.jwtSecret)
    if err != nil {
        return "", err
    }
    return tokenString, nil
}


func (auth *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate the signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrInvalidToken
        }
        return auth.jwtSecret, nil
    })

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    // Extract and validate claims
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }
    return nil, ErrInvalidToken
}


func (auth *AuthService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func (auth *AuthService) CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}


func (auth *AuthService) GetUser(r *http.Request) dto.AuthUser {
    user := r.Context().Value("user").(jwt.Claims)
    strId, _ := user.GetSubject()
    id, _  := strconv.ParseUint(strId, 10, 64)


    return dto.AuthUser{ID: id } 
}





