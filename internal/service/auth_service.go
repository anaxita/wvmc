package service

import (
	"context"
	"fmt"

	"github.com/anaxita/wvmc/internal/dal"
)

type Auth struct {
	repo *dal.UserRepository // TODO: user auth repo here.
}

func NewAuthService(repo *dal.UserRepository) *Auth {
	return &Auth{repo: repo}
}

// CreateRefreshToken create refresh token.
func (s *Auth) CreateRefreshToken(ctx context.Context, userID int64, refreshToken string) error {
	if err := s.repo.CreateRefreshToken(ctx, userID, refreshToken); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

// RefreshToken get refresh token.
func (s *Auth) RefreshToken(ctx context.Context, refreshToken string) error {
	err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("get refresh token: %w", err)
	}

	return nil
}
