package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
)

// Success writes a success response in the standard envelope. It returns an
// error (always nil) so handlers can write `return response.Success(...)`,
// matching the error-returning handlerFunc signature.
func Success(c *gin.Context, status int, data any) error {
	c.JSON(status, envelope{Data: data})
	return nil
}

// Fail writes an *apperror.AppError as a response in the standard envelope.
func Fail(c *gin.Context, e *apperror.AppError) {
	c.JSON(e.HTTPStatus, envelope{Error: &errorBody{
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}})
}

func Internal(c *gin.Context) {
	Fail(c, apperror.New(http.StatusInternalServerError, apperror.CodeInternalServerError,
		"Something went wrong. Please try again later."))
}
