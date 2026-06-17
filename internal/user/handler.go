package user

import (
	"errors"
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

func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		if fields, ok := request.AsValidationErrors(err); ok {
			response.FailValidation(c, fields)
			return
		}
		response.Fail(c, http.StatusBadRequest, response.CodeJSONFormatInvalid,
			"Request body must be valid JSON")
		return
	}

	user, err := h.userSvc.Signup(c.Request.Context(), SignupParams(req))
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			response.Fail(c, http.StatusConflict, response.CodeEmailAlreadyExists, "Email already exists")
		default:
			slog.Error("internal server error", "error", err)
			response.Internal(c)
		}
		return
	}

	response.Success(c, http.StatusCreated, newUserResponse(user))
}
