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

// UserServers get servers by user id.
func (s *Server) UserServers(ctx context.Context, id uuid.UUID) ([]entity.Server, error) {
	servers, err := s.repo.FindByUser(ctx, id)
	if err != nil {
		return servers, fmt.Errorf("get servers with user id %s: %w", id, err)
	}

	return servers, nil
}

// SetUserServers add servers to user.
func (s *Server) SetUserServers(ctx context.Context, userID uuid.UUID, serversIDs []string) error {
	if err := s.repo.SetUserServers(ctx, userID, serversIDs); err != nil {
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

// CreateOrUpdate creates server.
func (s *Server) CreateOrUpdate(ctx context.Context, server entity.Server) error {
	err := s.repo.Upsert(ctx, server)
	if err != nil {
		return fmt.Errorf("upsert server: %w", err)
	}

	return nil
}

// FindByID get server by id.
func (s *Server) FindByID(ctx context.Context, id int64) (entity.Server, error) {
	server, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return server, fmt.Errorf("get server with id %d: %w", id, err)
	}

	return server, nil
}

// Control controls server.
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

// LoadServersFromHVs servers data.
func (s *Server) LoadServersFromHVs(ctx context.Context) error {
	servers, err := s.control.Servers(ctx)
	if err != nil {
		return fmt.Errorf("load servers from hvs: %w", err)
	}

	err = s.repo.Upsert(ctx, servers...)
	if err != nil {
		return fmt.Errorf("upsert servers: %w", err)
	}

	return nil
}
