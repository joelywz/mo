package database

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

// GlobalMiddleware introduces a database connection into the context for
// universal access. It's advisable to apply this at the topmost level
// of a route.
func GlobalMiddleware(db *bun.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			ctx = WithContext(ctx, db)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// TxMiddleware retrieves the database connection from the context,
// initiates a transaction, and then substitutes the original database
// connection in the context with this new transaction, using a
// common interface.
func TxMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			db, err := FromContext(ctx)

			if err != nil {
				return err
			}

			tx, err := db.BeginTx(ctx, nil)

			if err != nil {
				return err
			}

			ctx = WithContext(ctx, tx)
			c.SetRequest(c.Request().WithContext(ctx))

			if err := next(c); err != nil {
				txErr := tx.Rollback()

				if txErr != nil {
					err = errors.Join(err, txErr)
				}

				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}

			return nil
		}
	}
}
