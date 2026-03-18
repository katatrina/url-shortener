package request

import (
	"strconv"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

func registerCustomRules() {
	_ = validate.RegisterValidation("max_bytes", validateMaxBytes)
	_ = validate.RegisterValidation("short_code", validateShortCode)
	_ = validate.RegisterValidation("strong_pass", validateStrongPass)

	registerTranslation("max_bytes", "{0} is too long")
	registerTranslation("short_code", "{0} must contain only letters and numbers")
	registerTranslation("strong_pass", "{0} must include at least one uppercase, one lowercase, one number, and one special character")
}

// validateMaxBytes checks byte length (not rune length).
// Needed because bcrypt only uses the first 72 bytes.
func validateMaxBytes(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	param := fl.Param()
	limit, err := strconv.Atoi(param)
	if err != nil {
		return false
	}
	return len(field) <= limit
}

// validateShortCode checks if the given field value is a valid short code
// based on the internal short code algorithm.
func validateShortCode(fl validator.FieldLevel) bool {
	return shortcode.IsValid(fl.Field().String())
}

// validateStrongPass checks that the password contains at least one uppercase letter,
// one lowercase letter, one digit, and one special character.
func validateStrongPass(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case !unicode.IsLetter(ch) && !unicode.IsDigit(ch):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}
