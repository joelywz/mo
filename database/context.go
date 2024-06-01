package database

import (
	"context"
	"errors"

	"github.com/uptrace/bun"
)

type BunKey struct{}

var (
	ErrNoBunInContext = errors.New("no bun in context")
)

// WithContext returns a new context with a database connection.
func WithContext(ctx context.Context, bun bun.IDB) context.Context {
	return context.WithValue(ctx, BunKey{}, bun)
}

// FromContext retrieves the database connection from the context.
// Returns ErrNoBunInContext if the database connection is not found.
func FromContext(ctx context.Context) (bun.IDB, error) {
	bun, ok := ctx.Value(BunKey{}).(bun.IDB)

	if !ok {
		return nil, ErrNoBunInContext
	}
	return bun, nil
}
