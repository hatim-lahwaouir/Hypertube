package services

import (
    gomail "gopkg.in/mail.v2"
    "os"
     "fmt"
)


type UserManagementService  struct {
    ownerEmail string
    sendEmailPassword string
    jwtSecret []byte 

}


var cacheUserManagement *UserManagementService


func NewUserManagementService () *UserManagementService {

    if cacheUserManagement == nil {

        cacheUserManagement =  &UserManagementService{ownerEmail: os.Getenv("OWNER"), sendEmailPassword: os.Getenv("EMAIL_PASSWORD")  }
    } 
    return  cacheUserManagement
}


func (u *UserManagementService) SendChangeEmail(email string, code string) error {

    message := gomail.NewMessage()

    // Set email headers
    message.SetHeader("From", u.ownerEmail)
    message.SetHeader("To", email)
    message.SetHeader("Subject", "This is an email sent via Gomail and Gmail SMTP")

    // Set email body
    message.SetBody("text/plain", fmt.Sprintf("code for change email validation '%s'", code))

    // Set up the SMTP dialer
    dialer := gomail.NewDialer("smtp.gmail.com", 587, u.ownerEmail, u.sendEmailPassword)

    // Send the email
    if err := dialer.DialAndSend(message); err != nil {
        fmt.Println("Error:", err)
        return err 
    } 
    return nil
}






