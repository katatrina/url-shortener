package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
	"github.com/katatrina/url-shortener/internal/link"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/user"
)

// handlerFunc is a handler that returns errors instead of writing error
// responses itself. Success responses are still written by the handler
// (response.Success).
type handlerFunc func(c *gin.Context) error

// wrap adapts a handlerFunc to gin.HandlerFunc, routing every error to writeError.
func wrap(h handlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			writeError(c, err)
		}
	}
}

func writeError(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		response.Fail(c, appErr)
		return
	}

	switch {
	case errors.Is(err, user.ErrEmailExists):
		response.Fail(c, apperror.New(http.StatusConflict,
			apperror.CodeEmailAlreadyExists, "Email already exists"))
	case errors.Is(err, user.ErrCredentialsIncorrect):
		response.Fail(c, apperror.New(http.StatusUnauthorized,
			apperror.CodeCredentialsIncorrect, "Incorrect email or password"))
	case errors.Is(err, link.ErrSlugExists):
		response.Fail(c, apperror.New(http.StatusConflict,
			apperror.CodeSlugAlreadyExists, "Slug already exists"))
	default:
		slog.ErrorContext(c.Request.Context(), "unexpected error",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"error", err,
		)
		response.Internal(c)
	}
}
