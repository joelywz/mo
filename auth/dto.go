package auth

import "time"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	AuthUserID string  `json:"authUserId"`
	UserID     *string `json:"userId"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AuthUserID string  `json:"authUserId"`
	UserID     *string `json:"userId"`
}

type TokenResponse struct {
	RefreshToken  string    `json:"refreshToken"`
	AccessToken   string    `json:"accessToken"`
	RefreshExpiry time.Time `json:"refreshExpiry"`
	AccessExpiry  time.Time `json:"accessExpiry"`
}

type VerifyResponse struct {
	AuthUserID string  `json:"authUserId"`
	UserID     *string `json:"userId"`
}

type LinkRequest struct {
	AuthUserID string `json:"authUserId"`
	UserID     string `json:"userId"`
}
