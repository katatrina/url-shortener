package request

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/response"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// báo lỗi bằng tên json (email, fullName), không phải tên field
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})

	_ = validate.RegisterValidation("max_bytes", validateMaxBytes)
	_ = validate.RegisterValidation("strong_password", validateStrongPassword)
}

// max_bytes: đếm BYTE, không phải rune — bcrypt chỉ dùng 72 byte đầu.
func validateMaxBytes(fl validator.FieldLevel) bool {
	limit, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}
	return len(fl.Field().String()) <= limit
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	var upper, lower, digit, special bool
	for _, ch := range fl.Field().String() {
		switch {
		case unicode.IsUpper(ch):
			upper = true
		case unicode.IsLower(ch):
			lower = true
		case unicode.IsDigit(ch):
			digit = true
		case !unicode.IsLetter(ch) && !unicode.IsDigit(ch):
			special = true
		}
	}
	return upper && lower && digit && special
}

func AsValidationErrors(err error) ([]response.FieldError, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil, false
	}
	out := make([]response.FieldError, 0, len(ve))
	for _, e := range ve {
		out = append(out, response.FieldError{
			Field:   e.Field(),
			Code:    mapTag(e.Tag()),
			Message: messageFor(e),
		})
	}
	return out, true
}

func mapTag(tag string) response.FieldErrorCode {
	switch tag {
	case "required":
		return response.FieldCodeRequired
	case "email":
		return response.FieldCodeInvalidFormat
	case "min", "gte":
		return response.FieldCodeTooShort
	case "max", "lte", "max_bytes":
		return response.FieldCodeTooLong
	case "strong_password":
		return response.FieldCodeWeakPassword
	default:
		return response.FieldCodeInvalid
	}
}

// tag của validator -> template. {field} = tên json, {param} = tham số rule.
//
//nolint:gosec // G101 false positive: đây là template thông báo lỗi, không phải credential.
var messages = map[string]string{
	"required":        "{field} is required",
	"email":           "{field} must be a valid email address",
	"min":             "{field} must be at least {param} characters",
	"max":             "{field} must be at most {param} characters",
	"max_bytes":       "{field} is too long",
	"strong_password": "{field} must include an uppercase letter, a lowercase letter, a number, and a special character",
}

func messageFor(e validator.FieldError) string {
	tmpl, ok := messages[e.Tag()]
	if !ok {
		tmpl = "{field} is invalid"
	}
	m := strings.ReplaceAll(tmpl, "{field}", e.Field())
	return strings.ReplaceAll(m, "{param}", e.Param())
}
