package service

import (
	"context"
	"fmt"

	"github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/google/uuid"
)

type Server struct {
	repo    *dal.ServerRepository
	control *Control
}

func NewServerService(repo *dal.ServerRepository, control *Control) *Server {
	return &Server{
		repo:    repo,
		control: control,
	}
}

// FindByUserID get servers by user id.
func (s *Server) FindByUserID(ctx context.Context, id uuid.UUID) ([]entity.Server, error) {
	servers, err := s.repo.FindByUser(ctx, id)
	if err != nil {
		return servers, fmt.Errorf("get servers with user id %s: %w", id, err)
	}

	return servers, nil
}

// AddServersToUser add servers to user.
func (s *Server) AddServersToUser(ctx context.Context, userID uuid.UUID, serversIDs []int64) error {
	if err := s.repo.AddServersToUser(ctx, userID, serversIDs); err != nil {
		return fmt.Errorf("add servers to user: %w", err)
	}

	return nil
}

// Servers return all servers.
func (s *Server) Servers(ctx context.Context) (servers []entity.Server, err error) {
	user, err := entity.CtxUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user from context: %w", err)
	}

	if user.Role == entity.RoleAdmin {
		servers, err = s.repo.Servers(ctx)
	} else {
		servers, err = s.repo.FindByUser(ctx, user.ID)
	}

	if err != nil {
		return nil, fmt.Errorf("get servers: %w", err)
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

// CreateOrUpdate creates server.
func (s *Server) CreateOrUpdate(ctx context.Context, server entity.Server) (entity.Server, error) {
	id, err := s.repo.Upsert(ctx, server)
	if err != nil {
		return server, fmt.Errorf("create server: %w", err)
	}

	server.ID = id

	return server, nil
}

// FindByID get server by id.
func (s *Server) FindByID(ctx context.Context, id int64) (entity.Server, error) {
	server, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return server, fmt.Errorf("get server with id %d: %w", id, err)
	}

	return server, nil
}

// Control control server.
func (s *Server) Control(ctx context.Context, serverID int64, command entity.Command) error {
	server, err := s.repo.FindByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("get server with id %d: %w", serverID, err)
	}

	if err := s.control.ControlServer(ctx, server, command); err != nil {
		return fmt.Errorf("control server: %w", err)
	}

	// TODO add notification

	return nil
}
