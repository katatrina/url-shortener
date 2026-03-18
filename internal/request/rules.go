package request

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

func registerCustomRules() {
	_ = validate.RegisterValidation("max_bytes", validateMaxBytes)
	_ = validate.RegisterValidation("short_code", validateShortCode)

	registerTranslation("max_bytes", "{0} is too long")
	registerTranslation("short_code", "{0} must contain only letters and numbers")
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
