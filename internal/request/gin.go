package request

import (
	"bytes"
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
//	    response.HandleJSONBindingError(c, err)
//	    return
//	}
func ShouldBindJSON(c *gin.Context, obj any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err = json.Unmarshal(body, obj); err != nil {
		return err
	}

	NormalizeStrings(obj)

	if err = validate.Struct(obj); err != nil {
		return err
	}

	return nil
}
