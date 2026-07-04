package user

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
)

type Handler struct {
	userSvc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{userSvc: svc}
}

func (h *Handler) Signup(c *gin.Context) error {
	var req SignupRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	user, err := h.userSvc.Signup(c.Request.Context(), SignupParams(req))
	if err != nil {
		return err
	}

	slog.InfoContext(c.Request.Context(), "user signed up",
		slog.String("user_id", user.ID),
	)

	return response.Success(c, http.StatusCreated, newUserResponse(user))
}

func (h *Handler) Login(c *gin.Context) error {
	var req LoginRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	result, err := h.userSvc.Login(c.Request.Context(), LoginParams(req))
	if err != nil {
		return err
	}

	slog.InfoContext(c.Request.Context(), "user logged in",
		slog.String("user_id", result.User.ID),
	)

	return response.Success(c, http.StatusOK, LoginResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.AccessTokenExpiresAt.Unix(),
		User:                 newUserResponse(result.User),
	})
}
