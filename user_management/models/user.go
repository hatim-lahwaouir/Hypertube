package models


import (
  "gorm.io/gorm"
 )




type UserEmail struct {
    gorm.Model
    Code  string
    Email string `gorm:"unique;not null"`
}
