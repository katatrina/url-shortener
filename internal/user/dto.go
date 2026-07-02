package user

type SignupRequest struct {
	Email    string `json:"email"    validate:"required,email" normalize:"trim,lower"`
	Password string `json:"password" validate:"required,min=8,max=32,max_bytes=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" normalize:"trim,lower"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken          string       `json:"accessToken"`
	AccessTokenExpiresAt int64        `json:"accessTokenExpiresAt"`
	User                 UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  *string `json:"fullName"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
}

func newUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}
