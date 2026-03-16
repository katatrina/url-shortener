package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/model"
)

func (h *Handler) ShortenURL(c *gin.Context) {
	var req ShortenURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	arg := model.ShortenURLParams{
		OriginalURL: req.OriginalURL,
		UserID:      middleware.GetAuthUserID(c),
	}

	if req.CustomAlias != nil {
		arg.CustomAlias = *req.CustomAlias
	}

	url, err := h.service.ShortenURL(c.Request.Context(), arg)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL format"})
		case errors.Is(err, model.ErrInvalidShortCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": "custom alias contains invalid characters (only a-z, A-Z, 0-9)"})
		case errors.Is(err, model.ErrShortCodeTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "custom alias is already taken"})
		default:
			log.Printf("[ERROR] failed to shorten URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, newURLResponse(url, h.baseURL))
}

func (h *Handler) ListUserURLs(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	urls, total, err := h.service.ListUserURLs(c.Request.Context(), userID, pageSize, offset)
	if err != nil {
		log.Printf("[ERROR] failed to list user URLs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]URLResponse, len(urls))
	for i := range urls {
		resp[i] = newURLResponse(&urls[i], h.baseURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     resp,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *Handler) GetUserURL(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	shortCode := c.Param("code")

	url, err := h.service.GetUserURL(c.Request.Context(), shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound),
			errors.Is(err, model.ErrURLOwnerMismatch):
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		default:
			log.Printf("[ERROR] failed to get URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, newURLResponse(url, h.baseURL))
}

func (h *Handler) DeleteUserURL(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	shortCode := c.Param("code")

	err := h.service.DeleteUserURL(c.Request.Context(), shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound),
			errors.Is(err, model.ErrURLOwnerMismatch):
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		default:
			log.Printf("[ERROR] failed to delete URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
