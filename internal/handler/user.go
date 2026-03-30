package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/response"
)

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)

	user, err := h.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			response.NotFound(c, response.CodeUserNotFound, "User not found")
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, newUserResponse(user), "User profile retrieved successfully")
}
