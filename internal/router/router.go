package router

import (
	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/link"
	"github.com/katatrina/url-shortener/internal/router/middleware"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/katatrina/url-shortener/internal/user"
)

func New(
	userHandler *user.Handler,
	linkHandler *link.Handler,
	tokenIssuer *token.Issuer,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Public redirect: slug -> destination. Deliberately minimal (Recovery only).
	// It's the hottest path, and browser navigation isn't subject to CORS.
	r.GET("/:slug", middleware.Recovery(), linkHandler.Redirect)

	// Everything under /v1 is the JSON API and gets the full middleware stack.
	v1 := r.Group("/v1")
	v1.Use(middleware.RequestID())
	v1.Use(middleware.AccessLog())
	v1.Use(middleware.CORS("http://localhost:5173")) // TODO: move allowed origins into config instead of hardcoding.
	v1.Use(middleware.Recovery())

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
