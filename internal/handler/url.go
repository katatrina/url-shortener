package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
)

func (h *Handler) ShortenURL(c *gin.Context) {
	var req ShortenURLRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		response.HandleJSONBindingError(c, err)
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
			response.BadRequest(c, response.CodeInvalidURL, "Invalid URL format")
		case errors.Is(err, model.ErrInvalidShortCode):
			response.BadRequest(c, response.CodeInvalidShortCode, "Custom alias contains invalid characters (only a-z, A-Z, 0-9)")
		case errors.Is(err, model.ErrShortCodeTaken):
			response.Conflict(c, response.CodeShortCodeTaken, "Custom alias is already taken")
		default:
			log.Printf("[ERROR] failed to shorten URL: %v", err)
			response.InternalServerError(c)
		}
		return
	}

	response.Created(c, newURLResponse(url, h.baseURL), "URL shortened successfully")
}

func (h *Handler) ListUserURLs(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	pagination := request.ParsePaginationParams(c)

	urls, total, err := h.service.ListUserURLs(c.Request.Context(), userID, pagination.Limit(), pagination.Offset())
	if err != nil {
		log.Printf("[ERROR] failed to list user URLs: %v", err)
		response.InternalServerError(c)
		return
	}

	resp := make([]URLResponse, len(urls))
	for i := range urls {
		resp[i] = newURLResponse(&urls[i], h.baseURL)
	}

	response.OKWithPagination(c, resp, "URLs retrieved successfully", pagination.Page, pagination.PageSize, total)
}

func (h *Handler) GetUserURL(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	shortCode := c.Param("code")

	url, err := h.service.GetUserURL(c.Request.Context(), shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound),
			errors.Is(err, model.ErrURLOwnerMismatch):
			response.NotFound(c, response.CodeURLNotFound, "URL not found")
		default:
			log.Printf("[ERROR] failed to get URL: %v", err)
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, newURLResponse(url, h.baseURL), "URL retrieved successfully")
}

func (h *Handler) DeleteUserURL(c *gin.Context) {
	userID := middleware.MustGetAuthUserID(c)
	shortCode := c.Param("code")

	err := h.service.DeleteUserURL(c.Request.Context(), shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrURLNotFound),
			errors.Is(err, model.ErrURLOwnerMismatch):
			response.NotFound(c, response.CodeURLNotFound, "URL not found")
		default:
			log.Printf("[ERROR] failed to delete URL: %v", err)
			response.InternalServerError(c)
		}
		return
	}

	response.NoContent(c)
}
