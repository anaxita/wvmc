package service

import (
	"context"
	"fmt"

	"github.com/anaxita/wvmc/internal/wvmc/dal"
	"github.com/anaxita/wvmc/internal/wvmc/entity"
)

type Server struct {
	repo *dal.ServerRepository
}

func NewServerService(repo *dal.ServerRepository) *Server {
	return &Server{repo: repo}
}

// FindByUserID get servers by user id.
func (s *Server) FindByUserID(ctx context.Context, userID int64) ([]entity.Server, error) {
	servers, err := s.repo.FindByUser(ctx, userID)
	if err != nil {
		return servers, fmt.Errorf("get servers with user id %d: %w", userID, err)
	}

	return servers, nil
}

// AddServersToUser add servers to user.
func (s *Server) AddServersToUser(ctx context.Context, userID int64, servers []entity.Server) error {
	if err := s.repo.AddServersToUser(ctx, userID, servers); err != nil {
		return fmt.Errorf("add servers to user: %w", err)
	}

	return nil
}

// Servers return all servers.
func (s *Server) Servers(ctx context.Context) ([]entity.Server, error) {
	servers, err := s.repo.Servers(ctx)
	if err != nil {
		return servers, fmt.Errorf("get servers: %w", err)
	}

	return servers, nil
}

// FindByTitle get server by title.
func (s *Server) FindByTitle(ctx context.Context, title string) (entity.Server, error) {
	server, err := s.repo.FindByTitle(ctx, title)
	if err != nil {
		return server, fmt.Errorf("get server with title %s: %w", title, err)
	}

	return server, nil
}

// FindByHvAndTitle get server by hv and title.
func (s *Server) FindByHvAndTitle(ctx context.Context, hv, title string) (entity.Server, error) {
	server, err := s.repo.FindByHvAndTitle(ctx, hv, title)
	if err != nil {
		return server, fmt.Errorf("get server with hv id %s and title %s: %w", hv, title, err)
	}

	return server, nil
}

// Create creates server.
func (s *Server) Create(ctx context.Context, server entity.Server) (entity.Server, error) {
	id, err := s.repo.Upsert(ctx, server)
	if err != nil {
		return server, fmt.Errorf("create server: %w", err)
	}

	server.ID = id

	return server, nil
}
