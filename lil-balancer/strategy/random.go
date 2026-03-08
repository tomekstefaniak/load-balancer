package strategy

import (
	"errors"
	"math/rand"
	"sync"

	"lil-balancer/config"
)

type Random struct {
	backends   *[]config.Backend
	backendsMu *sync.RWMutex
}

func NewRandom(backends *[]config.Backend, backendsMu *sync.RWMutex) *Random {
	return &Random{
		backends:   backends,
		backendsMu: backendsMu,
	}
}

func (r *Random) PickBackend() (config.Backend, error) {
	r.backendsMu.RLock()
	defer r.backendsMu.RUnlock()

	n := len(*r.backends)
	if n == 0 {
		return config.Backend{}, errors.New("no backends available")
	}

	return (*r.backends)[rand.Intn(n)], nil
}

func (r *Random) OnRelease(backend config.Backend) {}
