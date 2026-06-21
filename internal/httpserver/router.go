package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/httpserver/middleware"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/katatrina/url-shortener/internal/user"
)

func NewRouter(db *pgxpool.Pool, tokenIssuer *token.Issuer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.AccessLog(), middleware.Recovery())

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, tokenIssuer)
	userHandler := user.NewHandler(userSvc)

	v1 := r.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/login", userHandler.Login)
	}

	return r
}
