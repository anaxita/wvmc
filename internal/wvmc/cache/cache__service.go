package cache

import (
	"sync"

	"github.com/anaxita/wvmc/internal/wvmc/model"
)

type CacheService struct {
	mu      sync.RWMutex
	servers []model.Server
}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (c *CacheService) Servers() []model.Server {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.servers) == 0 {
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

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			c.servers[i].State = string(state)

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

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			c.servers[i].Network = string(state)

			break
		}
	}
}
