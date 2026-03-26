package handler

import (
	"github.com/katatrina/url-shortener/internal/service"
)

type Handler struct {
	service *service.Service
	baseURL string
}

func New(service *service.Service, baseURL string) *Handler {
	return &Handler{
		service: service,
		baseURL: baseURL,
	}
}
