package handler

import "github.com/katatrina/url-shortener/internal/model"

// -- Requests --

type ShortenURLRequest struct {
	LongURL     string `json:"longUrl" validate:"required,http_url,max=2048" normalize:"trim"`
	CustomAlias string `json:"customAlias" validate:"omitempty,short_code,min=3,max=30" normalize:"trim"`
	ExpiresAt   *int64 `json:"expiresAt" validate:"omitempty,gt=0"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=255" normalize:"trim,lower"`
	FullName    string `json:"fullName" validate:"required,min=2,max=100" normalize:"trim,single_space"`
	Password    string `json:"password" validate:"required,min=8,max_bytes=72,strong_pass"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" normalize:"trim,lower"`
	Password string `json:"password" validate:"required"`
}

// -- Responses --

type URLResponse struct {
	ShortCode  string `json:"shortCode"`
	ShortURL   string `json:"shortUrl"`
	LongURL    string `json:"longUrl"`
	ClickCount int64  `json:"clickCount"`
	ExpiresAt  *int64 `json:"expiresAt"`
	CreatedAt  int64  `json:"createdAt"`
}

type LoginResponse struct {
	AccessToken          string       `json:"accessToken"`
	AccessTokenExpiresAt int64        `json:"accessTokenExpiresAt"`
	User                 UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"fullName"`
	CreatedAt   int64  `json:"createdAt"`
}

// -- Converters --

func newURLResponse(u *model.URL, baseURL string) URLResponse {
	resp := URLResponse{
		ShortCode:  u.ShortCode,
		ShortURL:   baseURL + "/" + u.ShortCode,
		LongURL:    u.LongURL,
		ClickCount: u.ClickCount,
		CreatedAt:  u.CreatedAt.Unix(),
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
		FullName:    u.FullName,
		CreatedAt:   u.CreatedAt.Unix(),
	}
}
