package cache

import (
	"sync"

	"github.com/anaxita/wvmc/internal/wvmc/entity"
)

type CacheService struct {
	mu      sync.RWMutex
	servers []entity.Server
}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (c *CacheService) Servers() []entity.Server {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.servers) == 0 {
		return nil
	}

	return c.servers
}

func (c *CacheService) SetServers(s []entity.Server) {
	c.mu.Lock()
	c.servers = s
	c.mu.Unlock()
}

func (c *CacheService) SetServerState(s entity.Server, state entity.ServerState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			c.servers[i].State = string(state)

			break
		}
	}
}

func (c *CacheService) SetServerNetwork(s entity.Server, state entity.ServerState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			c.servers[i].Network = string(state)

			break
		}
	}
}
