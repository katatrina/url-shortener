package request

import (

	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func registerCustomMessages() {
	messages := map[string]string{
		"required":      "{0} is required",
		"required_with": "{0} is required when {1} is present",
		"email":         "{0} must be a valid email address",
		"min":           "{0} must be at least {1} characters",
		"max":           "{0} must be at most {1} characters",
		"gte":           "{0} must be at least {1}",
		"lte":           "{0} must be at most {1}",
	}

	for tag, msg := range messages {
		registerTranslation(tag, msg)
	}
}

func registerTranslation(tag, message string) {
	_ = validate.RegisterTranslation(tag, trans,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			fieldName := toCamelCase(fe.Field())
			paramValue := formatParams(fe.Param())
			t, _ := ut.T(tag, fieldName, paramValue)
			return t
		},
	)
}

// formatParams formats validation parameters to camelCase.
// Handles multiple field names separated by spaces (e.g., "DistrictCode WardCode").
func formatParams(param string) string {
	if param == "" {
		return param
	}

	// Split by space to handle multiple field names
	fields := strings.Fields(param)

	// Convert each field to camelCase (reuses existing toCamelCase from validator.go)
	camelFields := make([]string, len(fields))
	for i, field := range fields {
		camelFields[i] = toCamelCase(field)
	}

	// Join with "or" for better readability
	if len(camelFields) > 1 {
		return strings.Join(camelFields, " or ")
	}

	return camelFields[0]
}
