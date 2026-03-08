package strategy

import (
	"errors"
	"fmt"
	"sync"

	"lil-balancer/config"
)

type LeastConnections struct {
	backends    *[]config.Backend
	backendsMu  *sync.RWMutex
	connections map[string]int
	connMu      sync.Mutex
}

func NewLeastConnections(backends *[]config.Backend, backendsMu *sync.RWMutex) *LeastConnections {
	return &LeastConnections{
		backends:    backends,
		backendsMu:  backendsMu,
		connections: make(map[string]int),
	}
}

func (l *LeastConnections) PickBackend() (config.Backend, error) {
	l.backendsMu.RLock()
	defer l.backendsMu.RUnlock()
	l.connMu.Lock()
	defer l.connMu.Unlock()

	if len(*l.backends) == 0 {
		return config.Backend{}, errors.New("no backends available")
	}

	var picked config.Backend
	pickedConns := -1
	found := false

	for _, b := range *l.backends {
		key := backendKey(b)
		conns := l.connections[key]
		if !found || conns < pickedConns {
			picked = b
			pickedConns = conns
			found = true
		}
	}

	l.connections[backendKey(picked)]++
	return picked, nil
}

func (l *LeastConnections) OnRelease(backend config.Backend) {
	l.connMu.Lock()
	defer l.connMu.Unlock()

	key := backendKey(backend)
	if l.connections[key] > 0 {
		l.connections[key]--
	}
}

func backendKey(b config.Backend) string {
	return fmt.Sprintf("%s:%d", b.Address, b.Port)
}
