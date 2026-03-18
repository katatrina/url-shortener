package request

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

func registerCustomRules() {
	_ = validate.RegisterValidation("maxbytes", validateMaxBytes)
	_ = validate.RegisterValidation("shortcode", validateShortCode)

	registerTranslation("maxbytes", "{0} is too long")
	registerTranslation("shortcode", "{0} must contain only letters and numbers")
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
