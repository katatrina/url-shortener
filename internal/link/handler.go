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

func (h *Handler) CreateLink(c *gin.Context) error {
	var req CreateLinkRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	link, err := h.linkSvc.CreateLink(c.Request.Context(), CreateLinkParams{
		UserID:         middleware.UserID(c),
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c.Request.Context(), "link created",
		slog.String("link_id", link.ID),
		slog.String("slug", link.Slug),
		slog.String("user_id", link.UserID),
	)

	return response.Success(c, http.StatusCreated, newLinkResponse(link))
}
