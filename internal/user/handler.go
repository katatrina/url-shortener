package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
)

type Handler struct {
	userService *Service
}

func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		response.HandleJSONBindingError(c, err)
		return
	}

	user, err := h.userService.Signup(c.Request.Context(), SignupParams(req))
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			response.Conflict(c, response.CodeEmailAlreadyExists, "Email already exists")
		default:
			response.InternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusCreated, newUserResponse(user))
}
