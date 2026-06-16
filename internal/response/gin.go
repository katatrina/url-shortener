package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/katatrina/url-shortener/internal/request"
)

// builder .
func builder(c *gin.Context) *Builder {
	return New()
}

func OK(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, builder(c).Success(data, message).Build())
}

func Created(c *gin.Context, data any, message string) {
	c.JSON(http.StatusCreated, builder(c).Success(data, message).Build())
}

func OKWithPagination(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, builder(c).Success(data, message))
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, code ErrorCode, message string) {
	c.JSON(http.StatusBadRequest, builder(c).Error(code, message).Build())
}

func BadRequestWithErrors(c *gin.Context, code ErrorCode, message string, errors []request.FieldError) {
	c.JSON(http.StatusBadRequest, builder(c).Error(code, message).WithErrors(errors).Build())
}

func Gone(c *gin.Context, code ErrorCode, message string) {
	c.JSON(http.StatusGone, builder(c).Error(code, message).Build())
}

func Unauthorized(c *gin.Context, code ErrorCode, message string) {
	c.JSON(http.StatusUnauthorized, builder(c).Error(code, message).Build())
}

func NotFound(c *gin.Context, code ErrorCode, message string) {
	c.JSON(http.StatusNotFound, builder(c).Error(code, message).Build())
}

func Conflict(c *gin.Context, code ErrorCode, message string) {
	c.JSON(http.StatusConflict, builder(c).Error(code, message).Build())
}

func InternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError,
		builder(c).Error(CodeInternalServerError, "Internal server error. Please try again later").Build(),
	)
}

func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, builder(c).Error(CodeTooManyRequests, message).Build())
}

// HandleJSONBindingError properly handles different types of request.ShouldBindJSON errors.
// It distinguishes between JSON parsing errors and validation errors,
// returning appropriate error codes and messages.
//
// Usage:
//
//	if err := request.ShouldBindJSON(&req); err != nil {
//	    response.HandleJSONBindingError(c, err)
//	    return
//	}
func HandleJSONBindingError(c *gin.Context, err error) {
	var validationErrors validatorV10.ValidationErrors
	if errors.As(err, &validationErrors) {
		fieldErrors := request.TranslateValidationErrors(validationErrors)
		BadRequestWithErrors(c, CodeValidationFailed, "Validation failed", fieldErrors)
		return
	}

	BadRequest(c, CodeJSONFormatInvalid, "Request body must be valid JSON")
}
