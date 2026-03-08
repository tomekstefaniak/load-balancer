package strategy

import (
	"errors"
	"sync"

	"lil-balancer/config"
)

type RoundRobin struct {
	backends    *[]config.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
	current     int
}

func NewRoundRobin(backends *[]config.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections) *RoundRobin {
	return &RoundRobin{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
	}
}

func (rr *RoundRobin) PickBackend() (config.Backend, error) {
	rr.backendsMu.RLock()
	defer rr.backendsMu.RUnlock()
	rr.connTracker.Mu.Lock()
	defer rr.connTracker.Mu.Unlock()

	n := len(*rr.backends)
	if n == 0 {
		return config.Backend{}, errors.New("no backends available")
	}

	for i := 0; i < n; i++ {
		idx := (rr.current + i) % n
		b := (*rr.backends)[idx]
		key := BackendKey(b)
		if rr.connTracker.Conns[key] < b.MaxConnections {
			rr.current = (idx + 1) % n
			rr.connTracker.Conns[key]++
			return b, nil
		}
	}

	return config.Backend{}, errors.New("all backends at max connections")
}

func (rr *RoundRobin) OnRelease(backend config.Backend) {
	rr.connTracker.Mu.Lock()
	defer rr.connTracker.Mu.Unlock()

	key := BackendKey(backend)
	if rr.connTracker.Conns[key] > 0 {
		rr.connTracker.Conns[key]--
	}
}
