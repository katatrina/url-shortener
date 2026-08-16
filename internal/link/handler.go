package link

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/apperror"
	"github.com/katatrina/url-shortener/internal/click"
	"github.com/katatrina/url-shortener/internal/request"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/router/middleware"
)

type ClickRecorder interface {
	Record(e click.Event)
}

type Handler struct {
	linkSvc       *Service
	shortURLBase  string
	clickRecorder ClickRecorder
}

func NewHandler(svc *Service, shortURLBase string, clickRecorder ClickRecorder) *Handler {
	return &Handler{
		linkSvc:       svc,
		shortURLBase:  shortURLBase,
		clickRecorder: clickRecorder,
	}
}

func (h *Handler) CreateLink(c *gin.Context) error {
	var req CreateLinkRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	link, err := h.linkSvc.CreateLink(c.Request.Context(), CreateLinkParams{
		UserID:         middleware.UserID(c),
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
		Slug:           req.Slug,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c.Request.Context(), "link created",
		slog.String("link_id", link.ID),
		slog.String("slug", link.Slug),
		slog.String("user_id", link.UserID),
	)

	return response.Success(c, http.StatusCreated, newLinkResponse(link, h.shortURLBase))
}

func (h *Handler) GetLinkStats(c *gin.Context) error {
	id := c.Param("id")

	if err := uuid.Validate(id); err != nil {
		return ErrLinkNotFound
	}

	rng, err := parseStatsRange(c.Query("range"))
	if err != nil {
		return err
	}

	loc, err := parseTimezone(c.Query("tz"))
	if err != nil {
		return err
	}

	stats, err := h.linkSvc.GetLinkStats(c.Request.Context(), GetLinkStatsParams{
		LinkID:   id,
		UserID:   middleware.UserID(c),
		Range:    rng,
		Location: loc,
	})
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, newLinkStatsResponse(stats, rng, loc))
}

func parseStatsRange(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultStatsRange, nil
	}

	switch raw {
	case RangeLast24Hours, RangeLast7Days, RangeLast30Days, RangeLast90Days:
		return raw, nil
	}

	return "", apperror.New(http.StatusUnprocessableEntity, apperror.CodeValidationFailed,
		"Validation failed", apperror.FieldError{
			Field:   "range",
			Code:    apperror.FieldCodeInvalid,
			Message: "range must be one of: 24h, 7d, 30d, 90d",
		})
}

func parseTimezone(raw string) (*time.Location, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.UTC, nil
	}

	invalid := apperror.New(http.StatusUnprocessableEntity, apperror.CodeValidationFailed,
		"Validation failed", apperror.FieldError{
			Field:   "tz",
			Code:    apperror.FieldCodeInvalid,
			Message: "tz must be a valid IANA timezone name, e.g. Asia/Ho_Chi_Minh",
		})

	if len(raw) > maxTimezoneLen {
		return nil, invalid
	}

	loc, err := time.LoadLocation(raw)
	if err != nil {
		return nil, invalid
	}

	return loc, nil
}

func (h *Handler) ListLinks(c *gin.Context) error {
	links, err := h.linkSvc.ListLinks(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, newListLinksResponse(links, h.shortURLBase))
}

func (h *Handler) UpdateLink(c *gin.Context) error {
	id := c.Param("id")

	if err := uuid.Validate(id); err != nil {
		return ErrLinkNotFound
	}

	var req UpdateLinkRequest
	if err := request.ShouldBindJSON(c, &req); err != nil {
		return err
	}

	if req.IsEmpty() {
		return apperror.New(http.StatusUnprocessableEntity, apperror.CodeValidationFailed,
			"At least one field must be provided with a non-null value")
	}

	link, err := h.linkSvc.UpdateLink(c.Request.Context(), UpdateLinkParams{
		ID:             id,
		UserID:         middleware.UserID(c),
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
	})
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, newLinkResponse(link, h.shortURLBase))
}

func (h *Handler) DeleteLink(c *gin.Context) error {
	id := c.Param("id")

	if err := uuid.Validate(id); err != nil {
		return ErrLinkNotFound
	}

	if err := h.linkSvc.DeleteLink(c.Request.Context(), id, middleware.UserID(c)); err != nil {
		return err
	}

	return response.NoContent(c)
}

func (h *Handler) Redirect(c *gin.Context) {
	rawSlug := c.Param("slug")

	link, err := h.linkSvc.ResolveSlug(c.Request.Context(), rawSlug)
	if err != nil {
		if errors.Is(err, ErrLinkNotFound) {
			c.String(http.StatusNotFound, "Link not found")
			return
		}

		slog.ErrorContext(c.Request.Context(), "redirect lookup failed",
			slog.String("slug", rawSlug),
			slog.Any("error", err),
		)
		c.String(http.StatusInternalServerError, "Something went wrong")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, link.DestinationURL)

	e := click.NewEvent(link.ID, c.ClientIP(), c.Request.Referer(), c.Request.UserAgent())
	h.clickRecorder.Record(e)
}
