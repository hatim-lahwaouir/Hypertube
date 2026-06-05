package services

import (
    gomail "gopkg.in/mail.v2"
    "os"
    "time"
    "math/rand"
    "strconv"
    "fmt"

)



func SendCodeViaMail(email string, code string) error {
    var (
        owner string
        password string

    )
    owner = os.Getenv("OWNER")
    password = os.Getenv("EMAIL_PASSWORD")

    message := gomail.NewMessage()
    

    // Set email headers
    message.SetHeader("From", owner)
    message.SetHeader("To", email)
    message.SetHeader("Subject", "This is an email sent via Gomail and Gmail SMTP")

    // Set email body
    message.SetBody("text/plain", fmt.Sprintf("code for email validation '%s'", code))

    // Set up the SMTP dialer
    dialer := gomail.NewDialer("smtp.gmail.com", 587, owner, password)

    // Send the email
    if err := dialer.DialAndSend(message); err != nil {
        fmt.Println("Error:", err)
        return err 
    } 
    return nil
}




func GenerateOneTimeCode() string {
    var (
        nbr int
    )


    rand.Seed(time.Now().UnixNano())
    min := 999999 
    max := 9999999 
    nbr = rand.Intn(max - min + 1) + min
    return strconv.Itoa(nbr)
}
