package request

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

// FieldErrorCode represents the type of validation error.
type FieldErrorCode string

const (
	FieldCodeRequired      FieldErrorCode = "REQUIRED"
	FieldCodeInvalidFormat FieldErrorCode = "INVALID_FORMAT"
	FieldCodeMinValue      FieldErrorCode = "MIN_VALUE"
	FieldCodeMaxValue      FieldErrorCode = "MAX_VALUE"
	FieldCodeWeakPassword  FieldErrorCode = "WEAK_PASSWORD"
)

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string         `json:"field"`
	Value   any            `json:"value"`
	Code    FieldErrorCode `json:"code"`
	Message string         `json:"message"`
}

func init() {
	validate = validator.New()

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ = uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(validate, trans)

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return field.Name
		}
		return name
	})

	registerCustomMessages()
	registerCustomRules()
}

// TranslateValidationErrors converts validator.ValidationErrors to a slice of FieldError.
func TranslateValidationErrors(errs validator.ValidationErrors) []FieldError {
	fieldErrors := make([]FieldError, 0, len(errs))
	for _, e := range errs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   toCamelCase(e.Field()),
			Value:   e.Value(),
			Code:    mapValidationTag(e.Tag()),
			Message: e.Translate(trans),
		})
	}

	return fieldErrors
}

// mapValidationTag maps validator tags to FieldErrorCode constants.
func mapValidationTag(tag string) FieldErrorCode {
	switch tag {
	case "required":
		return FieldCodeRequired
	case "email":
		return FieldCodeInvalidFormat
	case "min", "gte":
		return FieldCodeMinValue
	case "max", "lte":
		return FieldCodeMaxValue
	case "strong_pass":
		return FieldCodeWeakPassword

	default:
		return FieldErrorCode(strings.ToUpper(tag))
	}
}

// toCamelCase converts PascalCase to camelCase.
// DisplayName -> displayName
// UserID -> userId (properly handles acronyms)
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Convert the first character to the lowercase
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])

	// Handle acronyms: UserID -> userId, not userID
	// Find where the acronym ends (when we hit a lowercase letter)
	for i := 1; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i+1]) {
			// This is where acronym ends
			// e.g., "UserID" -> i=4 (D), so convert I to lowercase
			// Result: "userId"
			break
		}
		if unicode.IsUpper(runes[i]) {
			runes[i] = unicode.ToLower(runes[i])
		} else {
			break
		}
	}

	return string(runes)
}
