package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
)

// maxBodyBytes caps the request body size to avoid loading a huge body into memory (OOM/DoS).
const maxBodyBytes = 1 << 20 // 1MB

// ShouldBindJSON caps the body size, reads, unmarshals JSON, normalizes,
// then validates obj.
//
// Returns nil on success. Errors that "know how to be returned to the client"
// (body too large, malformed JSON, validation failure) come back as
// *apperror.AppError. Unknown errors are returned as-is — writeError logs
// them and responds 500.
//
// Usage in a handler (handlers return errors, they don't write error responses):
//
//	var req Request
//	if err := request.ShouldBindJSON(c, &req); err != nil {
//	    return err
//	}
func ShouldBindJSON(c *gin.Context, obj any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return apperror.New(http.StatusRequestEntityTooLarge, apperror.CodePayloadTooLarge,
				"Request body is too large")
		}
		return err
	}

	if err := json.Unmarshal(body, obj); err != nil {
		// Valid JSON but one field has the wrong type (e.g. number into a
		// string field). encoding/json stops at the first error, so there is
		// exactly 1 field -> return a field error through the same 422
		// envelope as validation. Field == "" means a top-level type mismatch
		// (body is not an object) -> fall through to the generic message.
		if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok && typeErr.Field != "" {
			return apperror.New(http.StatusUnprocessableEntity, apperror.CodeValidationFailed,
				"Validation failed", apperror.FieldError{
					Field:   typeErr.Field,
					Code:    apperror.FieldCodeInvalidFormat,
					Message: typeErr.Field + " has an invalid type",
				})
		}
		return apperror.New(http.StatusBadRequest, apperror.CodeJSONFormatInvalid,
			"Request body must be a JSON object")
	}

	NormalizeStrings(obj)

	if err := validate.Struct(obj); err != nil {
		if fields, ok := AsValidationErrors(err); ok {
			return apperror.New(http.StatusUnprocessableEntity,
				apperror.CodeValidationFailed, "Validation failed", fields...)
		}
		return err
	}

	return nil
}
