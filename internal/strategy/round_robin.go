package strategy

import (
	"errors"
	"sync"

	cmn "load-balancer/internal/common"
)

type RoundRobinStrategy struct {
	backends    []cmn.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
	connMu      *sync.Mutex
	current     int
}

func NewRoundRobin(backends []cmn.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections, connMu *sync.Mutex) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
		connMu:      connMu,
	}
}

func (rr *RoundRobinStrategy) PickBackend() (cmn.Backend, error) {
	rr.backendsMu.RLock()
	defer rr.backendsMu.RUnlock()
	rr.connMu.Lock()
	defer rr.connMu.Unlock()

	n := len(rr.backends)
	if n == 0 {
		return cmn.Backend{}, errors.New("no backends available")
	}

	for i := 0; i < n; i++ {
		idx := (rr.current + i) % n
		b := rr.backends[idx]
		key := BackendKey(b)
		if rr.connTracker.Conns[key] < b.MaxConnections {
			rr.current = (idx + 1) % n
			rr.connTracker.Conns[key]++
			return b, nil
		}
	}

	return cmn.Backend{}, errors.New("all backends at max connections")
}

func (rr *RoundRobinStrategy) OnRelease(backend cmn.Backend) {
	rr.connMu.Lock()
	defer rr.connMu.Unlock()

	key := BackendKey(backend)
	if rr.connTracker.Conns[key] > 0 {
		rr.connTracker.Conns[key]--
	}
}
