package handler

import (
	"github.com/katatrina/url-shortener/internal/analytics"
	"github.com/katatrina/url-shortener/internal/service"
)

type Handler struct {
	service   *service.Service
	collector *analytics.Collector
	baseURL   string
}

func New(service *service.Service, collector *analytics.Collector, baseURL string) *Handler {
	return &Handler{
		service:   service,
		collector: collector,
		baseURL:   baseURL,
	}
}
