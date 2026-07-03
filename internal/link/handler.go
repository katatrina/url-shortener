package link

import (
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

func (h *Handler) CreateLink(c *gin.Context) error {
	var req CreateLinkRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	created, err := h.linkSvc.CreateLink(c.Request.Context(), CreateLinkParams{
		OwnerID:        middleware.UserID(c),
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
	})
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusCreated, newLinkResponse(created))
}
