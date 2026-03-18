package response

type ErrorCode string

const (
	CodeSuccess ErrorCode = "OK"

	CodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	CodeJSONFormatInvalid ErrorCode = "INVALID_JSON_FORMAT"

	CodeURLNotFound    ErrorCode = "URL_NOT_FOUND"
	CodeURLExpired     ErrorCode = "URL_EXPIRED"
	CodeShortCodeTaken ErrorCode = "SHORT_CODE_TAKEN"

	CodeAuthRequired       ErrorCode = "AUTHENTICATION_REQUIRED"
	CodeCredentialsInvalid ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"

	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"

	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	CodeRouteNotFound       ErrorCode = "ROUTE_NOT_FOUND"
	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)
