package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
)

func Success(c *gin.Context, status int, data any) error {
	c.JSON(status, envelope{Data: data})
	return nil
}

func NoContent(c *gin.Context) error {
	c.Status(http.StatusNoContent)
	return nil
}

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
