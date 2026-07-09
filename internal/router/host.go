package router

import (
	"net"
	"net/http"
	"strings"
)

type hostRouter struct {
	redirectHost string
	redirect     http.Handler
	api          http.Handler
}

func (h *hostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if normalizeHost(r.Host) == h.redirectHost {
		h.redirect.ServeHTTP(w, r)
		return
	}
	h.api.ServeHTTP(w, r)
}

func normalizeHost(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.ToLower(host)
}
