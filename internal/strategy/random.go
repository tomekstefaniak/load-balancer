package strategy

import (
	"errors"
	"math/rand"
	"sync"

	cmn "load-balancer/internal/common"
)

type RandomStrategy struct {
	backends    *[]cmn.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
	connMu      *sync.Mutex
}

func NewRandom(backends *[]cmn.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections, connMu *sync.Mutex) *RandomStrategy {
	return &RandomStrategy{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
		connMu:      connMu,
	}
}

func (r *RandomStrategy) PickBackend() (cmn.Backend, error) {
	r.backendsMu.RLock()
	defer r.backendsMu.RUnlock()
	r.connMu.Lock()
	defer r.connMu.Unlock()

	n := len(*r.backends)
	if n == 0 {
		return cmn.Backend{}, errors.New("no backends available")
	}

	start := rand.Intn(n)
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		b := (*r.backends)[idx]
		key := BackendKey(b)
		if r.connTracker.Conns[key] < b.MaxConnections {
			r.connTracker.Conns[key]++
			return b, nil
		}
	}

	return cmn.Backend{}, errors.New("all backends at max connections")
}

func (r *RandomStrategy) OnRelease(backend cmn.Backend) {
	r.connMu.Lock()
	defer r.connMu.Unlock()

	key := BackendKey(backend)
	if r.connTracker.Conns[key] > 0 {
		r.connTracker.Conns[key]--
	}
}
