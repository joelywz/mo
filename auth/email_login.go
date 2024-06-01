package auth

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var _ bun.BeforeAppendModelHook = (*EmailLogin)(nil)

type EmailLogin struct {
	bun.BaseModel `bun:"email_logins"`
	Email         string    `bun:"email,pk,notnull,type:varchar(320)"`
	Password      string    `bun:"password,notnull,type:varchar(128)"`
	AuthUserID    string    `bun:"auth_user_id,type:varchar(32)"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
	UpdatedAt     time.Time `bun:"updated_at,notnull"`
}

// BeforeAppendModel implements schema.BeforeAppendModelHook.
func (e *EmailLogin) BeforeAppendModel(ctx context.Context, query schema.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		e.CreatedAt = time.Now()
		e.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		e.UpdatedAt = time.Now()
	}

	return nil
}
