// Package apperror defines the shared "error vocabulary" of the transport
// layer: error codes, field errors, and AppError (an error that already
// knows how to be returned to the client).
//
// Dependency conventions:
//   - apperror does NOT import gin or any other package in this project.
//   - request produces *AppError (binding/validation errors).
//   - router/writeError maps service-layer sentinel errors -> responses.
//   - The service layer does NOT use this package — services return sentinel errors only.
package apperror

type ErrorCode string

const (
	CodeJSONFormatInvalid    ErrorCode = "INVALID_JSON_FORMAT"
	CodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	CodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	CodeEmailAlreadyExists   ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeCredentialsIncorrect ErrorCode = "INCORRECT_CREDENTIALS"
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	CodePayloadTooLarge      ErrorCode = "PAYLOAD_TOO_LARGE"
	CodeSlugAlreadyExists    ErrorCode = "SLUG_ALREADY_EXISTS"
	CodeLinkQuotaExceeded    ErrorCode = "LINK_QUOTA_EXCEEDED"
)

type FieldErrorCode string

const (
	FieldCodeRequired      FieldErrorCode = "REQUIRED"
	FieldCodeInvalidFormat FieldErrorCode = "INVALID_FORMAT"
	FieldCodeTooShort      FieldErrorCode = "TOO_SHORT"
	FieldCodeTooLong       FieldErrorCode = "TOO_LONG"
	// FieldCodeWeakPassword FieldErrorCode = "WEAK_PASSWORD" // disabled: relaxed password policy
	FieldCodeInvalid FieldErrorCode = "INVALID"
)

type FieldError struct {
	Field   string         `json:"field"`
	Code    FieldErrorCode `json:"code"`
	Message string         `json:"message"`
}

// AppError is a transport-layer error: it carries the HTTP status,
// error code, and a message that is safe to return to the client as-is.
type AppError struct {
	HTTPStatus int
	Code       ErrorCode
	Message    string
	Fields     []FieldError
}

func (e *AppError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func New(httpStatus int, code ErrorCode, message string, fields ...FieldError) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Fields:     fields,
	}
}
