package strategy

import (
	"errors"
	"sync"

	"lil-balancer/config"
)

type RoundRobin struct {
	backends   *[]config.Backend
	backendsMu *sync.RWMutex
	current    int
}

func NewRoundRobin(backends *[]config.Backend, backendsMu *sync.RWMutex) *RoundRobin {
	return &RoundRobin{
		backends:   backends,
		backendsMu: backendsMu,
	}
}

func (r *RoundRobin) PickBackend() (config.Backend, error) {
	r.backendsMu.RLock()
	defer r.backendsMu.RUnlock()

	n := len(*r.backends)
	if n == 0 {
		return config.Backend{}, errors.New("no backends available")
	}

	picked := (*r.backends)[r.current%n]
	r.current = (r.current + 1) % n
	return picked, nil
}

func (r *RoundRobin) OnRelease(backend config.Backend) {}
