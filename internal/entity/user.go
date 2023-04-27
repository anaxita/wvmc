package entity

import (
	"context"
)

type User struct {
	ID       int64    `json:"id" db:"id"`
	Name     string   `json:"name" db:"name"`
	Email    string   `json:"email" db:"email"`
	Company  string   `json:"company" db:"company"`
	Role     UserRole `json:"role" db:"role"`
	Password string   `json:"-" db:"password"`
}

type UserRole = int

const (
	UserRoleUser UserRole = iota
	UserRoleAdmin
)

type CtxUserKey struct{}

// CtxUser returns user from context
func CtxUser(ctx context.Context) (User, error) {
	user, ok := ctx.Value(CtxUserKey{}).(User)
	if !ok {
		return User{}, ErrUnauthorized
	}

	return user, nil
}
