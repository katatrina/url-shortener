package apperror

type ErrorCode string

const (
	CodeJSONFormatInvalid    ErrorCode = "INVALID_JSON_FORMAT"
	CodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	CodeEmailAlreadyExists   ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeCredentialsIncorrect ErrorCode = "INCORRECT_CREDENTIALS"
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	CodePayloadTooLarge      ErrorCode = "PAYLOAD_TOO_LARGE"
	CodeSlugAlreadyExists    ErrorCode = "SLUG_ALREADY_EXISTS"
	CodeLinkQuotaExceeded    ErrorCode = "LINK_QUOTA_EXCEEDED"
	CodeLinkNotFound         ErrorCode = "LINK_NOT_FOUND"
)

type FieldErrorCode string

const (
	FieldCodeRequired      FieldErrorCode = "REQUIRED"
	FieldCodeInvalidFormat FieldErrorCode = "INVALID_FORMAT"
	FieldCodeTooShort      FieldErrorCode = "TOO_SHORT"
	FieldCodeTooLong       FieldErrorCode = "TOO_LONG"
	FieldCodeInvalid       FieldErrorCode = "INVALID"
	// FieldCodeWeakPassword FieldErrorCode = "WEAK_PASSWORD" // disabled: relaxed password policy
)

type FieldError struct {
	Field   string         `json:"field"`
	Code    FieldErrorCode `json:"code"`
	Message string         `json:"message"`
}

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
