package handler

import "github.com/katatrina/url-shortener/internal/model"

// -- Requests --

type ShortenURLRequest struct {
	OriginalURL string  `json:"originalUrl" binding:"required,url"`
	CustomAlias *string `json:"customAlias,omitempty"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email,max=255"`
	DisplayName string `json:"displayName" binding:"required,min=2,max=100"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// -- Responses --

type URLResponse struct {
	ShortCode   string `json:"shortCode"`
	ShortURL    string `json:"shortUrl"`
	OriginalURL string `json:"originalUrl"`
	ClickCount  int64  `json:"clickCount"`
	ExpiresAt   *int64 `json:"expiresAt,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

type LoginResponse struct {
	AccessToken          string       `json:"accessToken"`
	AccessTokenExpiresAt int64        `json:"accessTokenExpiresAt"`
	User                 UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	CreatedAt   int64  `json:"createdAt"`
}

// -- Converters --

func newURLResponse(u *model.URL, baseURL string) URLResponse {
	resp := URLResponse{
		ShortCode:   u.ShortCode,
		ShortURL:    baseURL + "/" + u.ShortCode,
		OriginalURL: u.OriginalURL,
		ClickCount:  u.ClickCount,
		CreatedAt:   u.CreatedAt.Unix(),
	}

	if u.ExpiresAt != nil {
		ts := u.ExpiresAt.Unix()
		resp.ExpiresAt = &ts
	}

	return resp
}

func newUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt.Unix(),
	}
}
