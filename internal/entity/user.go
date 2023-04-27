package entity

import (
	"context"
)

type UserCreate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Company  string `json:"company"`
	Role     Role   `json:"role"`
	Password string `json:"password"`
}

type UserEdit struct {
	Name     string `json:"name"`
	Company  string `json:"company"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type User struct {
	ID       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Company  string `json:"company" db:"company"`
	Role     Role   `json:"role" db:"role"`
	Password string `json:"-" db:"password"`
}

type Role = int

const (
	RoleUser Role = iota
	RoleAdmin
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
