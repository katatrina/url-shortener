package request

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePaginationParams extracts page and pageSize from query parameters.
func ParsePaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))

	return NewPaginationParams(page, pageSize)
}

// ShouldBindJSON binds JSON request body to obj, normalizes it, then validates.
//
// This is a drop-in replacement for gin.Context.ShouldBindJSON() with auto-normalization.
//
// Usage:
//
//	var req RegisterRequest
//	if err := request.ShouldBindJSON(c, &req); err != nil {
//	    response.HandleJSONBindingError(c, err)
//	    return
//	}
func ShouldBindJSON(c *gin.Context, obj interface{}) error {
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
