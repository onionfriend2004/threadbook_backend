package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	_ = validate.RegisterValidation("password", validatePassword)
	_ = validate.RegisterValidation("username", validateUsername)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber {
		return false
	}

	allowedChars := regexp.MustCompile(`^[A-Za-z0-9!@#$%^&*()_+\-=[\]{}|;:,.<>?]+$`)
	return allowedChars.MatchString(password)
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	if len(username) < 3 || len(username) > 32 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	return matched
}
