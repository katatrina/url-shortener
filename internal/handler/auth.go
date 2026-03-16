package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/model"
)

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.Register(c.Request.Context(), model.CreateUserParams{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		default:
			log.Printf("[ERROR] failed to register user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, newUserResponse(user))
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.Login(c.Request.Context(), model.LoginUserParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrIncorrectCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect email or password"})
		default:
			log.Printf("[ERROR] failed to login: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.AccessTokenExpiresAt.Unix(),
		User:                 newUserResponse(result.User),
	})
}
