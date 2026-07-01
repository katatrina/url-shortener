package link

import "github.com/gin-gonic/gin"

type Handler struct {
	linkSvc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{linkSvc: svc}
}

func (h *Handler) CreateLink(c *gin.Context) {
	c.Status(201)
}
