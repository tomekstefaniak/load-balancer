package flags

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	cmn "load-balancer/internal/common"
	"load-balancer/internal/strategy"
)

type Flag[V any] struct {
	Value V
	Ok    bool
}

type Flags struct {
	ConfigPath        Flag[string]
	ListenerPort      Flag[int]
	ClientPort        Flag[int]
	Strategy          Flag[int]
	ServerConnTimeout Flag[int]
	Timeout           Flag[int]
	MaxConnections    Flag[int]
	Backends          Flag[[]cmn.Backend]
}

func ParseFlags() (*Flags, error) {
	args := os.Args[1:]
	flags := &Flags{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--config requires a Value")
			}
			i++
			flags.ConfigPath = Flag[string]{Value: args[i], Ok: true}

		case "--listener-port":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--listener-port requires a Value")
			}
			i++
			port, err := strconv.Atoi(args[i])
			if err != nil {
				return &Flags{}, fmt.Errorf("--listener-port invalid Value: %q", args[i])
			}
			flags.ListenerPort = Flag[int]{Value: port, Ok: true}

		case "--client-port":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--client-port requires a Value")
			}
			i++
			port, err := strconv.Atoi(args[i])
			if err != nil {
				return &Flags{}, fmt.Errorf("--client-port invalid Value: %q", args[i])
			}
			flags.ClientPort = Flag[int]{Value: port, Ok: true}

		case "--strategy":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--strategy requires a Value")
			}
			i++
			strategy, Ok := strategy.StrategyMap[strings.ToLower(args[i])]
			if !Ok {
				return &Flags{}, fmt.Errorf("--strategy unknown Value: %q", args[i])
			}
			flags.Strategy = Flag[int]{Value: strategy, Ok: true}

		case "--server-conn-timeout":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--server-conn-timeout requires a Value")
			}
			i++
			sct, err := strconv.Atoi(args[i])
			if err != nil {
				return &Flags{}, fmt.Errorf("--server-conn-timeout invalid Value: %q", args[i])
			}
			flags.ServerConnTimeout = Flag[int]{Value: sct, Ok: true}

		case "--timeout":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--timeout requires a Value")
			}
			i++
			timeout, err := strconv.Atoi(args[i])
			if err != nil {
				return &Flags{}, fmt.Errorf("--timeout invalid Value: %q", args[i])
			}
			flags.Timeout = Flag[int]{Value: timeout, Ok: true}

		case "--max-connections":
			if i+1 >= len(args) {
				return &Flags{}, fmt.Errorf("--max-connections requires a Value")
			}
			i++
			maxConn, err := strconv.Atoi(args[i])
			if err != nil {
				return &Flags{}, fmt.Errorf("--max-connections invalid Value: %q", args[i])
			}
			flags.MaxConnections = Flag[int]{Value: maxConn, Ok: true}

		case "--backends":
			var backends []cmn.Backend
			for i+3 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				address := args[i+1]
				port, err := strconv.Atoi(args[i+2])
				if err != nil {
					return &Flags{}, fmt.Errorf("--backends invalid port: %q", args[i+2])
				}
				maxConn, err := strconv.Atoi(args[i+3])
				if err != nil {
					return &Flags{}, fmt.Errorf("--backends invalid max connections: %q", args[i+3])
				}
				backends = append(backends, cmn.Backend{Address: address, Port: port, MaxConnections: maxConn})
				i += 3
			}
			if len(backends) == 0 {
				return &Flags{}, fmt.Errorf("--backends requires at least one <address> <port> pair")
			}
			flags.Backends = Flag[[]cmn.Backend]{Value: backends, Ok: true}

		default:
			return &Flags{}, fmt.Errorf("unknown flag: %q", args[i])
		}
	}

	return flags, nil
}
