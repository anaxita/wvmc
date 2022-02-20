package cache

import (
	"fmt"
	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"sync"
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
	logit.Info(fmt.Printf("Меняем статус сервера ID %d NAME %s HV %s на %s", s.ID, s.Name, s.HV,
		state))
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			logit.Info("Успешно сменили статус")

			c.servers[i].State = string(state)

			break
		}
	}
}

func (c *CacheService) SetServerNetwork(s model.Server, state model.ServerState) {
	logit.Info(fmt.Printf("Меняем сеть сервера ID %d NAME %s HV %s на %s", s.ID, s.Name, s.HV,
		state))

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servers == nil {
		return
	}

	for i, v := range c.servers {
		if v.Name == s.Name &&
			v.HV == s.HV {
			logit.Info("Успешно сменили сеть")

			c.servers[i].Network = string(state)

			break
		}
	}
}
