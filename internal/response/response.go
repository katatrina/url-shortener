package response

type FieldError struct {
	Field   string         `json:"field"`
	Code    FieldErrorCode `json:"code"`
	Message string         `json:"message"`
}

type errorBody struct {
	Code    ErrorCode    `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}
