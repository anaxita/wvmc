package cache

import (
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"sync"
)

type CacheService struct {
	mu      sync.Mutex
	servers []model.Server
}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (c *CacheService) Servers() []model.Server {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.servers) < 1 {
		return nil
	}

	return c.servers
}

func (c *CacheService) SetServers(s []model.Server) {
	c.mu.Lock()
	c.servers = s
	c.mu.Unlock()
}

func (c *CacheService) SetServerState(s model.Server, state model.ServerState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for _, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			v.State = string(state)

			break
		}
	}
}

func (c *CacheService) SetServerNetwork(s model.Server, state model.ServerState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for _, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			v.Network = string(state)

			break
		}
	}
}
