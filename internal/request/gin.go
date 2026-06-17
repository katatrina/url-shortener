package request

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

// ShouldBindJSON binds JSON request body to obj, normalizes it, then validates.
//
// This is a drop-in replacement for gin.Context.ShouldBindJSON() with auto-normalization.
//
// Usage:
//
//	var req Request
//	if err := request.ShouldBindJSON(c, &req); err != nil {
//	    if fields, ok := request.AsValidationErrors(err); ok {
//	        response.FailValidation(c, fields)
//	        return
//	    }
//	    response.Fail(c, http.StatusBadRequest, response.CodeJSONFormatInvalid,
//	        "Request body must be valid JSON")
//	    return
//	}
func ShouldBindJSON(c *gin.Context, obj any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
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
