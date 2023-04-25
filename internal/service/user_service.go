package service

import (
	"context"
	"fmt"

	"github.com/anaxita/wvmc/internal/wvmc/dal"
	"github.com/anaxita/wvmc/internal/wvmc/entity"
	"github.com/anaxita/wvmc/pkg/hasher"
)

type User struct {
	repo *dal.UserRepository
}

func NewUserService(repo *dal.UserRepository) *User {
	return &User{repo: repo}
}

// FindByEmail get user by email.
func (s *User) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return user, fmt.Errorf("get user with email %s: %w", email, err)
	}

	return user, nil
}

// FindByID get user by id.
func (s *User) FindByID(ctx context.Context, id int64) (entity.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return user, fmt.Errorf("get user with id %d: %w", id, err)
	}

	return user, nil
}

// Users return all users.
func (s *User) Users(ctx context.Context) ([]entity.User, error) {
	users, err := s.repo.Users(ctx)
	if err != nil {
		return users, fmt.Errorf("get users: %w", err)
	}

	return users, nil
}

// Delete delete user by id.
func (s *User) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user with id %d: %w", id, err)
	}

	if err = s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user with id %d: %w", id, err)
	}

	return nil
}

// Create creates user.
func (s *User) Create(ctx context.Context, user entity.User) (entity.User, error) {
	hashedPassword, err := hasher.Hash(user.Password)
	if err != nil {
		return user, fmt.Errorf("hash password: %w", err)
	}

	user.Password = string(hashedPassword)

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return user, fmt.Errorf("create user: %w", err)
	}

	user.ID = id

	return user, nil
}

// Edit edits user.
func (s *User) Edit(ctx context.Context, user entity.User, withPass bool) error {
	_, err := s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("get user with id %d: %w", user.ID, err)
	}

	if err = s.repo.Edit(ctx, user, withPass); err != nil {
		return fmt.Errorf("edit user with id %d: %w", user.ID, err)
	}

	return nil
}
