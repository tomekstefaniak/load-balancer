package strategy

import (
	"errors"
	"sync"

	"lil-balancer/config"
)

type LeastConnections struct {
	backends    *[]config.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
}

func NewLeastConnections(backends *[]config.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections) *LeastConnections {
	return &LeastConnections{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
	}
}

func (lc *LeastConnections) PickBackend() (config.Backend, error) {
	lc.backendsMu.RLock()
	defer lc.backendsMu.RUnlock()
	lc.connTracker.Mu.Lock()
	defer lc.connTracker.Mu.Unlock()

	if len(*lc.backends) == 0 {
		return config.Backend{}, errors.New("no backends available")
	}

	var picked config.Backend
	pickedConns := -1
	found := false

	for _, b := range *lc.backends {
		key := BackendKey(b)
		conns := lc.connTracker.Conns[key]
		if conns >= b.MaxConnections {
			continue
		}
		if !found || conns < pickedConns {
			picked = b
			pickedConns = conns
			found = true
		}
	}

	if !found {
		return config.Backend{}, errors.New("all backends at max connections")
	}

	lc.connTracker.Conns[BackendKey(picked)]++
	return picked, nil
}

func (lc *LeastConnections) OnRelease(backend config.Backend) {
	lc.connTracker.Mu.Lock()
	defer lc.connTracker.Mu.Unlock()

	key := BackendKey(backend)
	if lc.connTracker.Conns[key] > 0 {
		lc.connTracker.Conns[key]--
	}
}
