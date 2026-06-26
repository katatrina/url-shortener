package router

import (
	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/router/middleware"
	"github.com/katatrina/url-shortener/internal/user"
)

func New(userHandler *user.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.AccessLog(), middleware.Recovery())

	v1 := r.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/login", userHandler.Login)
	}

	return r
}
