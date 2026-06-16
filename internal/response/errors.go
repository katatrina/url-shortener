package response

type ErrorCode string

const (
	CodeSuccess ErrorCode = "OK"

	CodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	CodeJSONFormatInvalid ErrorCode = "INVALID_JSON_FORMAT"

	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"

	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)
