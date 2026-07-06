package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
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

// writeError is the ONLY place that maps errors -> HTTP responses.
// New endpoints with new business errors only need a new case here.
//
// It lives in package router (not response) because it must import the
// domain packages (user, link) to recognize sentinel errors — and domain
// handlers import response, so putting it there creates an import cycle.
func writeError(c *gin.Context, err error) {
	// 1. Transport-layer errors already packaged (binding, validation) -> write as-is.
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		response.Fail(c, appErr)
		return
	}

	// 2. Sentinel errors from the service layer -> map to status + code here.
	// Note: link.ErrSlugExists is not mapped here because the service swallows
	// it in the retry loop; if it escapes, retries are exhausted -> 500 is right.
	switch {
	case errors.Is(err, user.ErrEmailExists):
		response.Fail(c, apperror.New(http.StatusConflict,
			apperror.CodeEmailAlreadyExists, "Email already exists"))
	case errors.Is(err, user.ErrCredentialsIncorrect):
		response.Fail(c, apperror.New(http.StatusUnauthorized,
			apperror.CodeCredentialsIncorrect, "Incorrect email or password"))
	default:
		// 3. Unrecognized -> log server-side, return a generic 500.
		// NEVER expose err.Error() to the client.
		slog.ErrorContext(c.Request.Context(), "unexpected error",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"error", err,
		)
		response.Internal(c)
	}
}
