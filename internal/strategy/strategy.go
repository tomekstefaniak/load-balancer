package strategy

import (
	"fmt"
	"sync"

	"load-balancer/internal/config"
)

type Strategy interface {
	PickBackend() (config.Backend, error)
	OnRelease(backend config.Backend)
}

// BackendConnections is a shared connection tracker owned by the balancer
// All strategies receive a pointer to the same instance so connection counts
// persist across strategy swaps.
type BackendConnections struct {
	Mu    sync.Mutex
	Conns map[string]int
}

func NewBackendConnections() *BackendConnections {
	return &BackendConnections{
		Conns: make(map[string]int),
	}
}

func BackendKey(b config.Backend) string {
	return fmt.Sprintf("%s:%d", b.Address, b.Port)
}
