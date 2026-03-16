package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/model"
)

func (h *Handler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := h.service.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found"})
		case errors.Is(err, model.ErrURLExpired):
			c.JSON(http.StatusGone, gin.H{"error": "this short URL has expired"})
		default:
			log.Printf("[ERROR] failed to resolve short URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// 302 Found: temporary redirect.
	// Browser will always hit our server, so we can track every click.
	// If we used 301 (permanent), the browser would cache and skip us.
	c.Redirect(http.StatusFound, originalURL)
}
