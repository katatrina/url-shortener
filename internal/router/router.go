package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/link"
	"github.com/katatrina/url-shortener/internal/response"
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

	return &hostRouter{
		redirectHost: cfg.RedirectHost,
		redirect:     newRedirectEngine(cfg, linkHandler),
		api:          newAPIEngine(cfg, userHandler, linkHandler, tokenIssuer),
	}
}

func newRedirectEngine(cfg *config.Config, linkHandler *link.Handler) *gin.Engine {
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

	// Cross-cutting middleware at engine root so they also cover preflight
	// OPTIONS and any unmatched route: those fall to the NoRoute chain, where
	// group-level middleware does NOT run. Order (outermost first):
	//   RequestID  -> every log below carries request_id
	//   AccessLog  -> wraps Recovery so it logs the recovered 500 status
	//   CORS       -> aborts preflight 204 (still logged, it's inside AccessLog)
	//   Recovery   -> tightly wraps the business handlers
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS(cfg.AllowedOrigins...))
	r.Use(middleware.Recovery())

	// Unmatched routes -> JSON 404 in the standard envelope (instead of gin's
	// plain-text default). Runs at the end of the root-middleware chain, so it
	// still gets request_id, access log, and CORS headers. Genuine preflight is
	// short-circuited by CORS earlier and never reaches here.
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, apperror.New(http.StatusNotFound,
			apperror.CodeResourceNotFound, "Resource not found"))
	})

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
