package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/response"
)

// maxBodyBytes caps the request body size to avoid loading a huge body into memory (OOM/DoS).
const maxBodyBytes = 1 << 20 // 1MB

// ErrBodyTooLarge is returned when the request body exceeds maxBodyBytes.
var ErrBodyTooLarge = errors.New("request body too large")

// ShouldBindJSON is the low-level API: caps body size, reads, unmarshals JSON,
// normalizes, then validates obj, returning the raw error for the caller to handle.
//
// Most handlers should use BindJSON, which maps the error to a standard response.
func ShouldBindJSON(c *gin.Context, obj any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return ErrBodyTooLarge
		}
		return err
	}

	if err = json.Unmarshal(body, obj); err != nil {
		return err
	}

	NormalizeStrings(obj)

	if err = validate.Struct(obj); err != nil {
		return err
	}

	return nil
}

// BindJSON binds the JSON body into obj, writing a standard error response on failure.
// Returns true on success; false means an error response was written (caller just returns).
//
// Usage:
//
//	var req Request
//	if !request.BindJSON(c, &req) {
//	    return
//	}
func BindJSON(c *gin.Context, obj any) bool {
	err := ShouldBindJSON(c, obj)
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, ErrBodyTooLarge):
		response.Fail(c, http.StatusRequestEntityTooLarge, response.CodePayloadTooLarge,
			"Request body is too large")
	default:
		if fields, ok := AsValidationErrors(err); ok {
			response.FailValidation(c, fields)
			break
		}
		response.Fail(c, http.StatusBadRequest, response.CodeJSONFormatInvalid,
			"Request body must be valid JSON")
	}

	return false
}
