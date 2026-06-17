package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, envelope{Data: data})
}

func Fail(c *gin.Context, status int, code ErrorCode, message string) {
	c.JSON(status, envelope{Error: &errorBody{Code: code, Message: message}})
}

func FailValidation(c *gin.Context, fields []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, envelope{Error: &errorBody{
		Code: CodeValidationFailed, Message: "Validation failed", Fields: fields,
	}})
}

func Internal(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, CodeInternalServerError,
		"Something went wrong. Please try again later.")
}
