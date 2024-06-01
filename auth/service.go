package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joelywz/mo/database"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/matthewhartstonge/argon2"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrBadToken           = errors.New("bad token")
	ErrNotFound           = errors.New("user not found")
)

type Service struct {
	cfg *Config
}

func NewService(cfg *Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Login(ctx context.Context, dto *LoginRequest) (*LoginResponse, error) {

	db, err := database.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	// Check if email exist in database
	exists, err := db.NewSelect().
		Model((*EmailLogin)(nil)).
		Where("email = ?", dto.Email).
		Exists(ctx)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	var emailLogin EmailLogin

	err = db.NewSelect().
		Model(&emailLogin).
		Where("email = ?", dto.Email).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	match, err := argon2.VerifyEncoded([]byte(dto.Password), []byte(emailLogin.Password))

	if err != nil {
		return nil, err
	}

	if !match {
		return nil, ErrInvalidCredentials
	}

	// Get auth user that is associated with the email
	var user User

	err = db.NewSelect().Model(&user).Where("id = ?", emailLogin.AuthUserID).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AuthUserID: emailLogin.AuthUserID,
		UserID:     user.UserID,
	}, nil
}

func (s *Service) Register(ctx context.Context, dto *RegisterRequest) (*RegisterResponse, error) {

	db, err := database.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	// Check if email exist in database
	exists, err := db.NewSelect().
		Model((*EmailLogin)(nil)).
		Where("email = ?", dto.Email).
		Exists(ctx)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailExists
	}

	// Create new auth user
	user := User{
		ID:      gonanoid.Must(32),
		Version: gonanoid.Must(32),
		UserID:  nil,
	}

	_, err = db.NewInsert().
		Model(&user).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	// Create new email password
	argon := argon2.DefaultConfig()

	encoded, err := argon.HashEncoded([]byte(dto.Password))

	if err != nil {
		return nil, err
	}

	emailLogin := EmailLogin{
		Email:      dto.Email,
		Password:   string(encoded),
		AuthUserID: user.ID,
	}

	// Save email and password
	if _, err := db.NewInsert().Model(&emailLogin).Exec(ctx); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		AuthUserID: user.ID,
		UserID:     user.UserID,
	}, nil

}

func (s *Service) Link(ctx context.Context, dto *LinkRequest) error {
	db, err := database.FromContext(ctx)

	if err != nil {
		return err
	}

	exists, err := db.NewSelect().
		Model((*User)(nil)).
		Where("id = ?", dto.AuthUserID).
		Exists(ctx)

	if err != nil {
		return err
	}

	if !exists {
		return ErrNotFound
	}

	_, err = db.NewUpdate().
		Model((*User)(nil)).
		Where("id = ?", dto.AuthUserID).
		Set("user_id = ?", dto.UserID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Revoke(ctx context.Context, authUserId string) error {

	// Verify if authUserId exists in database
	db, err := database.FromContext(ctx)

	if err != nil {
		return err
	}

	exists, err := db.NewSelect().
		Model((*User)(nil)).
		Where("id = ?", authUserId).
		Exists(ctx)

	if err != nil {
		return err
	}

	if !exists {
		return ErrNotFound
	}

	// Update version of auth user
	_, err = db.NewUpdate().
		Model((*User)(nil)).
		Where("id = ?", authUserId).
		Set("version = ?", gonanoid.Must(32)).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Verify(ctx context.Context, token string, tokenType TokenType) (*VerifyResponse, error) {

	claims, err := s.parseJwt(token)

	if err != nil {
		return nil, errors.Join(err, ErrBadToken)
	}

	// Check token type
	if claims.Type != tokenType {
		return nil, ErrBadToken
	}

	db, err := database.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	// Get user from database
	exists, err := db.NewSelect().Model((*User)(nil)).Where("id = ?", claims.ID).Exists(ctx)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrNotFound
	}

	var user User

	err = db.NewSelect().Model(&user).Where("id = ?", claims.ID).Scan(ctx)

	if err != nil {
		return nil, err
	}

	// Compare version
	if user.Version != claims.Version {
		return nil, ErrBadToken
	}

	return &VerifyResponse{
		AuthUserID: user.ID,
		UserID:     user.UserID,
	}, nil
}

func (s *Service) CreateTokens(ctx context.Context, authUserId string) (*TokenResponse, error) {

	db, err := database.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	var user User

	err = db.NewSelect().Model(&user).Where("id = ?", authUserId).Scan(ctx)

	if err != nil {
		return nil, err
	}

	refreshToken, refreshClaims, err := s.createJwt(authUserId, user.Version, TokenTypeRefresh)

	if err != nil {
		return nil, err
	}

	refreshExp, _ := refreshClaims.GetExpirationTime()

	accessToken, accessClaims, err := s.createJwt(authUserId, user.Version, TokenTypeAccess)

	if err != nil {
		return nil, err
	}

	accessExp, _ := accessClaims.GetExpirationTime()

	return &TokenResponse{
		RefreshToken:  refreshToken,
		AccessToken:   accessToken,
		RefreshExpiry: refreshExp.Time,
		AccessExpiry:  accessExp.Time,
	}, nil
}

func (s *Service) createJwt(authUserId string, version string, tokenType TokenType) (string, jwt.Claims, error) {

	claims := TokenClaims{
		ID:      authUserId,
		Version: version,
		Type:    tokenType,
	}

	switch tokenType {
	case TokenTypeAccess:
		claims.ExpiresAt = jwt.NewNumericDate(
			time.Now().Add(time.Second * time.Duration(s.cfg.AccessLifetimeSeconds)),
		)
	case TokenTypeRefresh:
		claims.ExpiresAt = jwt.NewNumericDate(
			time.Now().Add(time.Second * time.Duration(s.cfg.RefreshLifetimeSeconds)),
		)
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))

	if err != nil {
		return "", nil, err
	}

	return signed, claims, nil
}

func (s *Service) parseJwt(token string) (*TokenClaims, error) {

	claims := &TokenClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
