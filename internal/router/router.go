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

	redirect := newRedirectEngine(cfg, linkHandler)
	api := newAPIEngine(cfg, userHandler, linkHandler, tokenIssuer)

	mux := http.NewServeMux()
	mux.Handle(cfg.RedirectHost+"/", redirect)
	mux.Handle("/", api)

	return normalizeHost(mux)
}

func newEngine(cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.TrustedPlatform = cfg.TrustedPlatform

	return r
}

func newRedirectEngine(cfg *config.Config, linkHandler *link.Handler) *gin.Engine {
	r := newEngine(cfg)
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
	r := newEngine(cfg)
	r.RedirectTrailingSlash = false

	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS(cfg.AllowedOrigins...))
	r.Use(middleware.Recovery())

	auth := r.Group("/auth")
	{
		auth.POST("/signup", wrap(userHandler.Signup))
		auth.POST("/login", wrap(userHandler.Login))
	}

	r.GET("/me", middleware.Auth(tokenIssuer), wrap(userHandler.GetMe))

	links := r.Group("/links")
	links.Use(middleware.Auth(tokenIssuer))
	{
		links.POST("", wrap(linkHandler.CreateLink))
		links.GET("", wrap(linkHandler.ListLinks))
		links.GET("/:id", wrap(linkHandler.GetLink))
		links.GET("/:id/stats", wrap(linkHandler.GetLinkStats))
		links.PATCH("/:id", wrap(linkHandler.UpdateLink))
		links.DELETE("/:id", wrap(linkHandler.DeleteLink))
	}

	return r
}

func normalizeHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = strings.ToLower(r.Host)
		next.ServeHTTP(w, r)
	})
}
