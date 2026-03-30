package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/response"
)

func (h *Handler) GetURLStats(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	shortCode := c.Param("code")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	// Clamp to a reasonable range
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}

	stats, err := h.service.GetURLStats(c.Request.Context(), model.GetURLStatsParams{
		UserID:    userID,
		ShortCode: shortCode,
		Days:      days,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound),
			errors.Is(err, model.ErrURLOwnerMismatch):
			response.NotFound(c, response.CodeURLNotFound, "URL not found")
		default:
			slog.Error("failed to get URL stats", "error", err)
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, stats, "URL stats retrieved successfully")
}
