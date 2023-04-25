package dal

import (
	"sync"

	"github.com/anaxita/wvmc/internal/entity"
)

type Cache struct {
	mu      sync.RWMutex
	servers []entity.Server
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Servers() []entity.Server {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.servers) == 0 {
		return nil
	}

	return c.servers
}

func (c *Cache) SetServers(s []entity.Server) {
	c.mu.Lock()
	c.servers = s
	c.mu.Unlock()
}

func (c *Cache) SetServerState(s entity.Server, state entity.ServerState) {
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

func (c *Cache) SetServerNetwork(s entity.Server, state entity.ServerState) {
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
