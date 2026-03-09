package strategy

import (
	"fmt"

	cmn "load-balancer/internal/common"
)

const (
	RoundRobin = iota
	LeastConnections
	Random
)

var StrategyMap = map[string]int{
	"roundrobin":       RoundRobin,
	"leastconnections": LeastConnections,
	"random":           Random,
}

type Strategy interface {
	PickBackend() (cmn.Backend, error)
	OnRelease(backend cmn.Backend)
}

// BackendConnections is a shared connection tracker owned by the balancer
// All strategies receive a pointer to the same instance so connection counts
// persist across strategy swaps.
type BackendConnections struct {
	Conns map[string]int
}

func NewBackendConnections() *BackendConnections {
	return &BackendConnections{
		Conns: make(map[string]int),
	}
}

func BackendKey(b cmn.Backend) string {
	return fmt.Sprintf("%s:%d", b.Address, b.Port)
}
