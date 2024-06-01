package auth

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	jwt.RegisteredClaims
	ID      string    `json:"id"`
	Type    TokenType `json:"type"`
	Version string    `json:"version"`
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "ACCESS"
	TokenTypeRefresh TokenType = "REFRESH"
)
