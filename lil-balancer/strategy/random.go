package strategy

import (
	"errors"
	"math/rand"
	"sync"

	"lil-balancer/config"
)

type Random struct {
	backends    *[]config.Backend
	backendsMu  *sync.RWMutex
	connTracker *BackendConnections
}

func NewRandom(backends *[]config.Backend, backendsMu *sync.RWMutex, connTracker *BackendConnections) *Random {
	return &Random{
		backends:    backends,
		backendsMu:  backendsMu,
		connTracker: connTracker,
	}
}

func (r *Random) PickBackend() (config.Backend, error) {
	r.backendsMu.RLock()
	defer r.backendsMu.RUnlock()
	r.connTracker.Mu.Lock()
	defer r.connTracker.Mu.Unlock()

	n := len(*r.backends)
	if n == 0 {
		return config.Backend{}, errors.New("no backends available")
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

	return config.Backend{}, errors.New("all backends at max connections")
}

func (r *Random) OnRelease(backend config.Backend) {
	r.connTracker.Mu.Lock()
	defer r.connTracker.Mu.Unlock()

	key := BackendKey(backend)
	if r.connTracker.Conns[key] > 0 {
		r.connTracker.Conns[key]--
	}
}
