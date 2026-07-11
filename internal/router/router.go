package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/link"
	"github.com/katatrina/url-shortener/internal/router/middleware"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/katatrina/url-shortener/internal/user"
)

func New(
	cfg *config.Config,
	userHandler *user.Handler,
	linkHandler *link.Handler,
	tokenIssuer *token.Issuer,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	redirect := newRedirectEngine(linkHandler)
	api := newAPIEngine(cfg, userHandler, linkHandler, tokenIssuer)

	mux := http.NewServeMux()
	mux.Handle(cfg.RedirectHost+"/", redirect)
	mux.Handle("/", api)

	return normalizeHost(mux)
}

func newRedirectEngine(linkHandler *link.Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())

	r.GET("/:slug", linkHandler.Redirect)

	return r
}

func newAPIEngine(
	cfg *config.Config,
	userHandler *user.Handler,
	linkHandler *link.Handler,
	tokenIssuer *token.Issuer,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS(cfg.AllowedOrigins...))
	r.Use(middleware.Recovery())

	v1 := r.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/signup", wrap(userHandler.Signup))
		auth.POST("/login", wrap(userHandler.Login))
	}

	links := v1.Group("/links")
	links.Use(middleware.Auth(tokenIssuer))
	{
		links.POST("", wrap(linkHandler.CreateLink))
	}

	return r
}

func normalizeHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = strings.ToLower(r.Host)
		next.ServeHTTP(w, r)
	})
}
