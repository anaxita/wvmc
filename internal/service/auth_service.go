package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/pkg/hasher"
	"github.com/dgrijalva/jwt-go"
)

const (
	tokenSubjectAccess  = "access"
	tokenSubjectRefresh = "refresh"
)

type Auth struct {
	repo *dal.UserRepository // TODO: user auth repo here.
}

func NewAuthService(repo *dal.UserRepository) *Auth {
	return &Auth{repo: repo}
}

// CreateRefreshToken create refresh token.
func (s *Auth) CreateRefreshToken(ctx context.Context, user entity.User) (string, error) {
	refreshToken := s.createRefreshToken(user)

	if err := s.repo.CreateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return "", fmt.Errorf("create refresh token: %w", err)
	}

	return refreshToken, nil
}

// SignIn sign in user.
func (s *Auth) SignIn(ctx context.Context, email, password string) (data entity.SignIn, err error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return data, fmt.Errorf("%w: find user by email: %s", entity.ErrUnauthorized, err)
	}

	err = hasher.Compare(user.Password, password)
	if err != nil {
		return data, fmt.Errorf("%w: compare password: %s", entity.ErrUnauthorized, err)
	}

	user.Password = ""

	refresh, err := s.CreateRefreshToken(ctx, user)
	if err != nil {
		return data, fmt.Errorf("%w: create refresh token: %s", entity.ErrUnauthorized, err)
	}

	access := s.createAccessToken(user)

	data = entity.SignIn{
		AccessToken:  access,
		RefreshToken: refresh,
		Role:         user.Role,
	}

	return data, nil
}

// RefreshTokens refresh tokens.
func (s *Auth) RefreshTokens(ctx context.Context, refreshToken string) (data entity.SignIn, err error) {
	err = s.repo.RefreshToken(ctx, refreshToken)
	if err != nil {
		return data, fmt.Errorf("%w: refresh token: %s", entity.ErrUnauthorized, err)
	}

	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return data, fmt.Errorf("%w: %s", entity.ErrUnauthorized, err)
	}

	if claims.Subject != tokenSubjectRefresh {
		return data, fmt.Errorf("%w: subject: %s", entity.ErrUnauthorized, claims.Subject)
	}

	data = entity.SignIn{
		AccessToken:  s.createAccessToken(claims.User),
		RefreshToken: s.createRefreshToken(claims.User),
		Role:         claims.User.Role,
	}

	return data, nil
}

type authClaims struct {
	jwt.StandardClaims
	User entity.User
}

func (s *Auth) createAccessToken(user entity.User) string {
	claims := authClaims{
		jwt.StandardClaims{
			Subject:   tokenSubjectAccess,
			ExpiresAt: time.Now().Add(time.Hour * 3600).Unix(), // TODO get from config.
		},
		user,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, _ := token.SignedString([]byte(os.Getenv("TOKEN"))) // TODO: get from config.

	return signedToken
}

func (s *Auth) createRefreshToken(user entity.User) string {
	claims := struct {
		jwt.StandardClaims
		User entity.User
	}{
		jwt.StandardClaims{
			Subject:   tokenSubjectRefresh,
			ExpiresAt: time.Now().Add(time.Hour * 3600).Unix(), // TODO get from config.
		},
		user,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, _ := token.SignedString([]byte(os.Getenv("TOKEN"))) // TODO: get from config.

	return signedToken
}

func (s *Auth) Auth(ctx context.Context, token string) (entity.User, error) {
	claims, err := s.parseToken(token)
	if err != nil {
		return entity.User{}, fmt.Errorf("parse token: %w", err)
	}

	if claims.Subject != tokenSubjectAccess {
		return entity.User{}, fmt.Errorf("%w: token subject is not valid", entity.ErrUnauthorized)
	}

	return claims.User, nil
}

func (s *Auth) parseToken(token string) (claims *authClaims, err error) {
	t, err := jwt.ParseWithClaims(token, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN")), nil // TODO: get from config.
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*authClaims)
	if !ok {
		return nil, fmt.Errorf("claims is not valid")
	}

	if !t.Valid {
		return nil, fmt.Errorf("expired at %d seconds", claims.ExpiresAt)
	}

	return claims, nil
}
