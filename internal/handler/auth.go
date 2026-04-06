package handler

import (
	"errors"
	"github.com/katatrina/url-shortener/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
)

func (h *Handler) Register(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	var req RegisterRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		response.HandleJSONBindingError(c, err)
		return
	}

	user, err := h.service.Register(c.Request.Context(), model.CreateUserParams{
		Email:       req.Email,
		FullName:    req.FullName,
		Password:    req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			response.Conflict(c, response.CodeEmailAlreadyExists, "Email already exists")
		default:
			log.Error("failed to register user", "error", err)
			response.InternalServerError(c)
		}
		return
	}

	response.Created(c, newUserResponse(user), "User registered successfully")
}

func (h *Handler) Login(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	var req LoginRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		response.HandleJSONBindingError(c, err)
		return
	}

	result, err := h.service.Login(c.Request.Context(), model.LoginUserParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrIncorrectCredentials):
			response.Unauthorized(c, response.CodeCredentialsInvalid, "Incorrect email or password")
		default:
			log.Error("failed to login", "error", err)
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, LoginResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.AccessTokenExpiresAt.Unix(),
		User:                 newUserResponse(result.User),
	}, "Login successfully")
}
