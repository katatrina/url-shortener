package request

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

func registerCustomRules() {
	_ = validate.RegisterValidation("maxbytes", validateMaxBytes)
	registerTranslation("maxbytes", "{0} is too long")
}

// validateMaxBytes checks byte length (not rune length)
// Needed because bcrypt only uses first 72 bytes
func validateMaxBytes(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	param := fl.Param()
	limit, err := strconv.Atoi(param)
	if err != nil {
		return false
	}
	return len(field) <= limit
}
