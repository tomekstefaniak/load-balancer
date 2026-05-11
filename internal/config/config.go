package config

import (
	"fmt"
	"os"
	"strings"

	cmn "load-balancer/internal/common"
	"load-balancer/internal/flags"
	"load-balancer/internal/strategy"

	"gopkg.in/yaml.v3"
)

var strategyNames = map[int]string{
	strategy.RoundRobin:       "roundrobin",
	strategy.LeastConnections: "leastconnections",
	strategy.Random:           "random",
}

func ParseStrategy(name string) (int, bool) {
	s, ok := strategy.StrategyMap[strings.ToLower(name)]
	return s, ok
}

func StrategyName(strategy int) string {
	if name, ok := strategyNames[strategy]; ok {
		return name
	}
	return "unknown"
}

type Config struct {
	ListenerPort          int
	ClientPort            int
	LoadBalancingStrategy int
	ServerConnTimeoutSec  int
	IdleTimeoutSec        int
	MaxConnections        int
	Backends              []cmn.Backend
}

type rawConfig struct {
	ListenerPort          int           `yaml:"ListenerPort"`
	ClientPort            int           `yaml:"ClientPort"`
	LoadBalancingStrategy string        `yaml:"LoadBalancingStrategy"`
	ServerConnTimeoutSec  int           `yaml:"ServerConnTimeoutSec"`
	IdleTimeoutSec        int           `yaml:"IdleTimeoutSec"`
	MaxConnections        int           `yaml:"MaxConnections"`
	Backends              []cmn.Backend `yaml:"Backends"`
}

func LoadConfig(f *flags.Flags) (*Config, error) {
	var raw rawConfig
	if f.ConfigPath.Ok {
		data, err := os.ReadFile(f.ConfigPath.Value)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	cfg := &Config{}

	// ListenerPort
	if f.ListenerPort.Ok {
		cfg.ListenerPort = f.ListenerPort.Value
	} else if f.ConfigPath.Ok {
		cfg.ListenerPort = raw.ListenerPort
	} else {
		return nil, fmt.Errorf("listener port not set: use --port or --config")
	}
	err := validatePort(cfg.ListenerPort)
	if err != nil {
		return nil, fmt.Errorf("invalid listener port: %w", err)
	}

	// ClientPort
	if f.ClientPort.Ok {
		cfg.ClientPort = f.ClientPort.Value
	} else if f.ConfigPath.Ok {
		cfg.ClientPort = raw.ClientPort
	} else {
		return nil, fmt.Errorf("client port not set: use --client-port or --config")
	}
	err = validatePort(cfg.ClientPort)
	if err != nil {
		return nil, fmt.Errorf("invalid client port: %w", err)
	}

	// LoadBalancingStrategy
	if f.Strategy.Ok {
		cfg.LoadBalancingStrategy = f.Strategy.Value
	} else if f.ConfigPath.Ok {
		strategy, ok := strategy.StrategyMap[strings.ToLower(raw.LoadBalancingStrategy)]
		if !ok {
			return nil, fmt.Errorf("unknown load balancing strategy: %q", raw.LoadBalancingStrategy)
		}
		cfg.LoadBalancingStrategy = strategy
	} else {
		return nil, fmt.Errorf("strategy not set: use --strategy or --config")
	}

	// ServerConnTimeoutSec
	if f.ServerConnTimeout.Ok {
		cfg.ServerConnTimeoutSec = f.ServerConnTimeout.Value
	} else if f.ConfigPath.Ok {
		cfg.ServerConnTimeoutSec = raw.ServerConnTimeoutSec
	} else {
		return nil, fmt.Errorf("server connection timeout not set: use --server-conn-timeout or --config")
	}
	if err := validateTimeoutSec(cfg.ServerConnTimeoutSec); err != nil {
		return nil, fmt.Errorf("invalid server connection timeout: %w", err)
	}

	// IdleTimeoutSec
	if f.Timeout.Ok {
		cfg.IdleTimeoutSec = f.Timeout.Value
	} else if f.ConfigPath.Ok {
		cfg.IdleTimeoutSec = raw.IdleTimeoutSec
	} else {
		return nil, fmt.Errorf("timeout not set: use --timeout or --config")
	}
	if err := validateTimeoutSec(cfg.IdleTimeoutSec); err != nil {
		return nil, fmt.Errorf("invalid idle timeout: %w", err)
	}

	// MaxConnections
	if f.MaxConnections.Ok {
		cfg.MaxConnections = f.MaxConnections.Value
	} else if f.ConfigPath.Ok {
		cfg.MaxConnections = raw.MaxConnections
	} else {
		return nil, fmt.Errorf("max connections not set: use --max-connections or --config")
	}

	// Backends - merge from both sources, deduplicate
	var backends []cmn.Backend
	if f.ConfigPath.Ok {
		backends = raw.Backends
	}
	if f.Backends.Ok {
		backends = mergeBackends(backends, f.Backends.Value)
	}
	if len(backends) == 0 {
		return nil, fmt.Errorf("at least one backend is required: use --backends or --config")
	}
	cfg.Backends = backends
	for _, b := range cfg.Backends {
		if err := validateBackend(b); err != nil {
			return nil, fmt.Errorf("invalid backend %s:%d: %w", b.Address, b.Port, err)
		}
	}

	return cfg, nil
}

func mergeBackends(backends []cmn.Backend, extraBackends []cmn.Backend) []cmn.Backend {
	seen := make(map[string]struct{})
	for _, b := range backends {
		key := fmt.Sprintf("%s:%d", b.Address, b.Port)
		seen[key] = struct{}{}
	}
	for _, b := range extraBackends {
		key := fmt.Sprintf("%s:%d", b.Address, b.Port)
		if _, ok := seen[key]; !ok {
			backends = append(backends, cmn.Backend{Address: b.Address, Port: b.Port, MaxConnections: b.MaxConnections})
			seen[key] = struct{}{}
		}
	}
	return backends
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func validateBackend(b cmn.Backend) error {
	if b.Address == "" {
		return fmt.Errorf("backend address cannot be empty")
	}
	if err := validatePort(b.Port); err != nil {
		return fmt.Errorf("invalid backend port: %w", err)
	}
	if b.MaxConnections <= 0 {
		return fmt.Errorf("backend max connections must be greater than 0")
	}
	return nil
}

func validateTimeoutSec(seconds int) error {
	if seconds < 1 {
		return fmt.Errorf("must be at least 1 second")
	}
	return nil
}
