package auth

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var _ bun.BeforeAppendModelHook = (*User)(nil)

type User struct {
	bun.BaseModel `bun:"auth_users"`
	ID            string    `bun:"id,pk,notnull,type:varchar(32)"`
	UserID        *string   `bun:"user_id,type:varchar(32)"`
	Version       string    `bun:"version,notnull,type:varchar(32)"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}

// BeforeAppendModel implements schema.BeforeAppendModelHook.
func (u *User) BeforeAppendModel(ctx context.Context, query schema.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now()
	}

	return nil
}
