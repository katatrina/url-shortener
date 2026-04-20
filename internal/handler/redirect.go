package handler

import (
	"errors"
	"net/http"

	"github.com/katatrina/url-shortener/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

func (h *Handler) Redirect(c *gin.Context) {
	log := logger.FromRequestContext(c.Request.Context())
	shortCode := c.Param("code")

	if !shortcode.IsValid(shortCode) {
		response.NotFound(c, response.CodeURLNotFound, "Short URL not found")
		return
	}

	destinationURL, err := h.service.ResolveAndTrack(c.Request.Context(), shortCode, model.ClickInfo{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound):
			response.NotFound(c, response.CodeURLNotFound, "Short URL not found")
		case errors.Is(err, model.ErrURLExpired):
			response.Gone(c, response.CodeURLExpired, "This short URL has expired")
		default:
			log.Error("failed to resolve short URL", "error", err)
			response.InternalServerError(c)
		}
		return
	}

	c.Redirect(http.StatusFound, destinationURL)
}
