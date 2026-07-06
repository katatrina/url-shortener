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

	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS("http://localhost:5173")) // TODO: move allowed origins into config instead of hardcoding.
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
