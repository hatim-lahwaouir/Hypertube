package dto
import (

    "github.com/go-playground/validator/v10"
    "unicode"
)

var Validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())







func strongPassword(password string) bool {
    var (
        hasUpper bool
        hasLower bool
        hasSpecial bool
        hasDegit bool
    )

    hasUpper = false
    hasLower = false
    hasSpecial = false
    hasDegit = false


	for _, r := range password {
		// Checks if the rune is not a letter, digit, or space
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			hasSpecial = true
		}

        if unicode.IsDigit(r) {
            hasDegit = true 
        }

        if unicode.IsUpper(r) {
            hasUpper = true
        }
        if unicode.IsLower(r) {
            hasLower = true
        }
    }


	return hasUpper && hasLower && hasSpecial && hasDegit 
}

func init() {
    // create a custom validation for password

    Validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(string)

        if strongPassword(value) == false {
            return false
        }

		return true 
	})


}
