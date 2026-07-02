package request

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	// "unicode" // disabled with validateStrongPassword

	"github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/slug"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})

	_ = validate.RegisterValidation("max_bytes", validateMaxBytes)
	// _ = validate.RegisterValidation("strong_password", validateStrongPassword) // disabled: relaxed password policy
	_ = validate.RegisterValidation("slug", validateSlug)
}

// max_bytes: đếm BYTE, không phải rune — bcrypt chỉ dùng 72 byte đầu.
func validateMaxBytes(fl validator.FieldLevel) bool {
	limit, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}
	return len(fl.Field().String()) <= limit
}

// validateStrongPassword: disabled — password policy relaxed to length-only (NIST 800-63B).
// Kept for easy restore; to re-enable also restore the "unicode" import,
// the RegisterValidation call, the mapTag case, and the messages entry.
// func validateStrongPassword(fl validator.FieldLevel) bool {
// 	var upper, lower, digit, special bool
// 	for _, ch := range fl.Field().String() {
// 		switch {
// 		case unicode.IsUpper(ch):
// 			upper = true
// 		case unicode.IsLower(ch):
// 			lower = true
// 		case unicode.IsDigit(ch):
// 			digit = true
// 		case !unicode.IsLetter(ch) && !unicode.IsDigit(ch):
// 			special = true
// 		}
// 	}
// 	return upper && lower && digit && special
// }

func validateSlug(fl validator.FieldLevel) bool {
	return slug.IsValid(fl.Field().String())
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
	case "email", "http_url":
		return response.FieldCodeInvalidFormat
	case "min", "gte":
		return response.FieldCodeTooShort
	case "max", "lte", "max_bytes":
		return response.FieldCodeTooLong
	// case "strong_password": // disabled: relaxed password policy
	// 	return response.FieldCodeWeakPassword
	case "slug":
		return response.FieldCodeInvalid
	default:
		return response.FieldCodeInvalid
	}
}

// G101 false positive: đây là template thông báo lỗi, không phải credential.
//
//nolint:gosec
var messages = map[string]string{
	"required":  "{field} is required",
	"email":     "{field} must be a valid email address",
	"http_url":  "{field} must be a valid http or https URL",
	"min":       "{field} must be at least {param} characters",
	"max":       "{field} must be at most {param} characters",
	"max_bytes": "{field} is too long",
	// "strong_password": "{field} must include an uppercase letter, a lowercase letter, a number, and a special character", // disabled: relaxed password policy
	"slug": "{field} must contain only letters, digits, '-' or '_'",
}

func messageFor(e validator.FieldError) string {
	tmpl, ok := messages[e.Tag()]
	if !ok {
		tmpl = "{field} is invalid"
	}
	m := strings.ReplaceAll(tmpl, "{field}", e.Field())
	return strings.ReplaceAll(m, "{param}", e.Param())
}
