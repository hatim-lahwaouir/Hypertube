package dto

import (
	"github.com/go-playground/validator/v10"
	"slices"
	"strconv"
	"time"
)

var Validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

var imdbGenres []string = []string{
	"Action",
	"Adult",
	"Adventure",
	"Animation",
	"Biography",
	"Comedy",
	"Crime",
	"Documentary",
	"Drama",
	"Family",
	"Fantasy",
	"Film-Noir",
	"Game-Show",
	"History",
	"Horror",
	"Music",
	"Musical",
	"Mystery",
	"News",
	"Reality-TV",
	"Romance",
	"Sci-Fi",
	"Short",
	"Sport",
	"Talk-Show",
	"Thriller",
	"War",
	"Western",
}

func init() {

	Validate.RegisterValidation("valid_genre", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(string)

		if slices.Contains(imdbGenres, value) == false {
			return false
		}

		return true
	})

	Validate.RegisterValidation("valid_year", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(string)
		year, err := strconv.Atoi(value)

		if err != nil {
			return false
		}

		if year > time.Now().Year() || year < 1900 {
			return false
		}
		return true
	})

}
