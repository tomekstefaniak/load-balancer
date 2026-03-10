package strategy

import (
	"errors"
	"sync"

	cmn "load-balancer/internal/common"
)

type LeastConnectionsStrategy struct {
	backends    *[]cmn.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
	connMu      *sync.Mutex
}

func NewLeastConnections(backends *[]cmn.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections, connMu *sync.Mutex) *LeastConnectionsStrategy {
	return &LeastConnectionsStrategy{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
		connMu:      connMu,
	}
}

func (lc *LeastConnectionsStrategy) PickBackend() (cmn.Backend, error) {
	lc.backendsMu.RLock()
	defer lc.backendsMu.RUnlock()
	lc.connMu.Lock()
	defer lc.connMu.Unlock()

	if len(*lc.backends) == 0 {
		return cmn.Backend{}, errors.New("no backends available")
	}

	var picked cmn.Backend
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
		return cmn.Backend{}, errors.New("all backends at max connections")
	}

	lc.connTracker.Conns[BackendKey(picked)]++
	return picked, nil
}

func (lc *LeastConnectionsStrategy) OnRelease(backend cmn.Backend) {
	lc.connMu.Lock()
	defer lc.connMu.Unlock()

	key := BackendKey(backend)
	if lc.connTracker.Conns[key] > 0 {
		lc.connTracker.Conns[key]--
	}
}
