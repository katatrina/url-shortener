package response

type ErrorCode string

const (
	CodeValidationFailed    ErrorCode = "VALIDATION_FAILED"
	CodeJSONFormatInvalid   ErrorCode = "INVALID_JSON_FORMAT"
	CodeEmailAlreadyExists  ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)

type FieldErrorCode string

const (
	FieldCodeRequired      FieldErrorCode = "REQUIRED"
	FieldCodeInvalidFormat FieldErrorCode = "INVALID_FORMAT"
	FieldCodeTooShort      FieldErrorCode = "TOO_SHORT"
	FieldCodeTooLong       FieldErrorCode = "TOO_LONG"
	FieldCodeWeakPassword  FieldErrorCode = "WEAK_PASSWORD"
	FieldCodeInvalid       FieldErrorCode = "INVALID"
)
