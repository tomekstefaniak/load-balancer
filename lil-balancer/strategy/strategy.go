package strategy

import (
	"lil-balancer/config"
)

type Strategy interface {
	PickBackend() (config.Backend, error)
	OnRelease(backend config.Backend)
}
