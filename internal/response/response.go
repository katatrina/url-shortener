package response

import "github.com/katatrina/url-shortener/internal/apperror"

type errorBody struct {
	Code    apperror.ErrorCode    `json:"code"`
	Message string                `json:"message"`
	Fields  []apperror.FieldError `json:"fields,omitempty"`
}

type envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}
