package service

import (
	"context"
	"fmt"

	"github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/pkg/hasher"
	"github.com/google/uuid"
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
func (s *User) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return user, fmt.Errorf("get user with id %s: %w", id, err)
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
func (s *User) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user with id %s: %w", id, err)
	}

	if err = s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user with id %s: %w", id, err)
	}

	return nil
}

// Create creates user.
func (s *User) Create(ctx context.Context, uc entity.UserCreate) (user entity.User, err error) {
	// TODO add normalize and validation

	uc.ID = uuid.New()

	hashedPassword, err := hasher.Hash(uc.Password)
	if err != nil {
		return user, fmt.Errorf("hash password: %w", err)
	}
	uc.Password = string(hashedPassword)

	err = s.repo.Create(ctx, uc)
	if err != nil {
		return user, fmt.Errorf("create user: %w", err)
	}

	user = entity.User{
		ID:      uc.ID,
		Email:   uc.Email,
		Name:    uc.Name,
		Company: uc.Company,
		Role:    uc.Role,
	}

	return user, nil
}

// Edit edits user.
func (s *User) Edit(ctx context.Context, id uuid.UUID, user entity.UserEdit) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user with id %s: %w", id, err)
	}

	if err = s.repo.Update(ctx, id, user); err != nil {
		return fmt.Errorf("edit user with id %d: %w", id, err)
	}

	return nil
}
