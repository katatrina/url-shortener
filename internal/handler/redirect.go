package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

func (h *Handler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	if !shortcode.IsValid(shortCode) {
		response.NotFound(c, response.CodeURLNotFound, "Short URL not found")
		return
	}

	originalURL, err := h.service.Resolve(c.Request.Context(), shortCode, model.ClickMeta{
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
			log.Printf("[ERROR] failed to resolve short URL: %v", err)
			response.InternalServerError(c)
		}
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}
