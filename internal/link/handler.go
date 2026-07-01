package link

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/router/middleware"
)

type Handler struct {
	linkSvc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{linkSvc: svc}
}

func (h *Handler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		if fields, ok := request.AsValidationErrors(err); ok {
			response.FailValidation(c, fields)
			return
		}
		response.Fail(c, http.StatusBadRequest, response.CodeJSONFormatInvalid,
			"Request body must be valid JSON")
		return
	}

	created, err := h.linkSvc.CreateLink(c.Request.Context(), CreateLinkParams{
		OwnerID:        middleware.UserID(c),
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to create link", "error", err)
		response.Internal(c)
		return
	}

	response.Success(c, http.StatusCreated, newLinkResponse(created))
}
