package auth_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joelywz/mo/auth"
	"github.com/joelywz/mo/database"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

var db *bun.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")

	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()

	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=root",
			"MYSQL_DATABASE=mo_auth",
			"MYSQL_DATABASE=mo_auth",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := resource.Expire(60); err != nil {
		log.Fatalf("Could not set expiry for resource: %s", err)
	}

	if err := pool.Retry(func() error {
		manager, err := database.NewMySQLManager(&database.Config{
			Dialect: "mysql",
			Host:    "localhost",
			Port:    resource.GetPort("3306/tcp"),
			User:    "root",
			Pass:    "root",
			Name:    "mo_auth",
		})

		if err != nil {
			return err
		}

		db, err = manager.Bun()

		if err != nil {
			return err
		}

		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// Migrate database
	if _, err := db.NewCreateTable().Model((*auth.User)(nil)).Exec(context.Background()); err != nil {
		log.Fatalf("Could not create table: %s", err)
	}

	if _, err := db.NewCreateTable().Model((*auth.EmailLogin)(nil)).Exec(context.Background()); err != nil {
		log.Fatalf("Could not create table: %s", err)

	}

	log.Println("Ready for testing")

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestService(t *testing.T) {

	authService := auth.NewService(&auth.Config{
		Secret:          "secret",
		RefreshDuration: 2 * time.Second,
		AccessDuration:  4 * time.Second,
	})

	email := "email@email.com"
	password := "1234567890"

	ctx := context.Background()
	ctx = database.WithContext(ctx, db)

	registerRes, err := authService.Register(ctx, &auth.RegisterRequest{
		Email:    email,
		Password: password,
	})

	assert.NoError(t, err, "register should not return error")

	// Registration
	_, err = authService.Register(ctx, &auth.RegisterRequest{
		Email:    email,
		Password: password,
	})

	assert.ErrorIs(t, err, auth.ErrEmailExists, "register should return ErrEmailExists")

	// Login
	_, err = authService.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})

	assert.NoError(t, err, "login should not return error")

	// Login with wrong password
	_, err = authService.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials, "login should return ErrInvalidCredentials")

	// Create tokens
	tokens, err := authService.CreateTokens(ctx, registerRes.AuthUserID)

	assert.NoError(t, err, "create tokens should not return error")

	// Token verification

	_, err = authService.Verify(ctx, tokens.AccessToken, auth.TokenTypeAccess)

	assert.NoError(t, err, "verify access token should not return error")

	// Token expiry and verification
	time.Sleep(time.Until(tokens.AccessExpiry))

	_, err = authService.Verify(ctx, tokens.AccessToken, auth.TokenTypeAccess)

	assert.ErrorIs(t, err, auth.ErrBadToken, "verify access token should return ErrBadToken after access expiry")

	_, err = authService.Verify(ctx, tokens.RefreshToken, auth.TokenTypeRefresh)

	assert.NoError(t, err, "verify refresh token should not return error")

	time.Sleep(time.Until(tokens.RefreshExpiry))

	_, err = authService.Verify(ctx, tokens.RefreshToken, auth.TokenTypeRefresh)

	assert.ErrorIs(t, err, auth.ErrBadToken, "verify refersh token should return ErrBadToken after refresh expiry ")

	// Token revocation
	tokens, err = authService.CreateTokens(ctx, registerRes.AuthUserID)

	assert.NoError(t, err, "create tokens should not return error")

	err = authService.Revoke(ctx, registerRes.AuthUserID)

	assert.NoError(t, err, "revoke should not return error")

	_, err = authService.Verify(ctx, tokens.AccessToken, auth.TokenTypeAccess)

	assert.ErrorIs(t, err, auth.ErrBadToken, "verify access token should return ErrBadToken after revocation")

	_, err = authService.Verify(ctx, tokens.AccessToken, auth.TokenTypeRefresh)

	assert.ErrorIs(t, err, auth.ErrBadToken, "verify refresh token should return ErrBadToken after revocation")

}
