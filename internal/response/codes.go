package response

type ErrorCode string

const (
	CodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	CodeJSONFormatInvalid    ErrorCode = "INVALID_JSON_FORMAT"
	CodeEmailAlreadyExists   ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeCredentialsIncorrect ErrorCode = "INCORRECT_CREDENTIALS"
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	CodePayloadTooLarge      ErrorCode = "PAYLOAD_TOO_LARGE"
)

type FieldErrorCode string

const (
	FieldCodeRequired      FieldErrorCode = "REQUIRED"
	FieldCodeInvalidFormat FieldErrorCode = "INVALID_FORMAT"
	FieldCodeTooShort      FieldErrorCode = "TOO_SHORT"
	FieldCodeTooLong       FieldErrorCode = "TOO_LONG"
	// FieldCodeWeakPassword  FieldErrorCode = "WEAK_PASSWORD" // disabled: relaxed password policy
	FieldCodeInvalid FieldErrorCode = "INVALID"
)
