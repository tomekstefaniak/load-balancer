package balancer

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"

	"lil-balancer/config"
	"lil-balancer/strategy"
)

const (
	BUSY = iota
	IDLE
)

type Balancer struct {
	Config                  *config.Config
	State 				    int
	StateMu				   *sync.RWMutex
	Strategy                strategy.Strategy
	LoadBalancingStrategyMu *sync.Mutex
	ConnectionTimeoutSecMu  *sync.RWMutex
	MaxConnectionsMu        *sync.Mutex
	BackendsMu              *sync.RWMutex
}

func NewBalancer(cfg *config.Config) *Balancer {
	var strat strategy.Strategy
	backendsMu := &sync.RWMutex{}
	switch cfg.LoadBalancingStrategy {
	case config.RoundRobin:
		// strat = strategy.NewRoundRobinStrategy(cfg.Backends, backendsMu)
	case config.LeastConnections:
		// strat = strategy.NewLeastConnectionsStrategy(cfg.Backends, backendsMu)
	case config.Random:
		// strat = strategy.NewRandomStrategy(cfg.Backends, backendsMu)
	default:
		panic(fmt.Sprintf("invalid load balancing strategy: %d", cfg.LoadBalancingStrategy))
	}

	return &Balancer{
		Config:                  cfg,
		State:                   IDLE,
		StateMu:                 &sync.RWMutex{},
		Strategy:                strat,
		LoadBalancingStrategyMu: &sync.Mutex{},
		ConnectionTimeoutSecMu:  &sync.RWMutex{},
		MaxConnectionsMu:        &sync.Mutex{},
		BackendsMu:              backendsMu,
	}
}

/* Load balancing logic */

func (b *Balancer) Balance(
	gracefulShutdownCtx context.Context,
	immediateShutdownCtx context.Context,
) {
	// State management
	b.StateMu.Lock()
	if b.State == BUSY {
		b.StateMu.Unlock()
		panic("balancer is already running")
	}
	b.State = BUSY
	b.StateMu.Unlock()
	defer func() { 
		b.StateMu.Lock()
		b.State = IDLE
		b.StateMu.Unlock()
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", b.Config.ListenerPort))
	if err != nil {
		panic(fmt.Errorf("failed to start listener: %w", err))
	}

	// Set up SIGTERM signal handling for graceful shutdown
	sigTermCtx, sigTermCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer sigTermCancel() // Close the context when balancing ends

	balancingWG := &sync.WaitGroup{}
	connectionsCtx, cancel := context.WithCancel(context.Background())

	// Accept incoming client connections
	balancingWG.Add(1)
	go func() {
		defer balancingWG.Done()
		for {
			clientConn, err := listener.Accept()
			if err != nil {
				select {
				case <-gracefulShutdownCtx.Done():
					return // Stop accepting new connections on shutdown
				case <-sigTermCtx.Done():
					return // Stop accepting new connections on shutdown
				case <-immediateShutdownCtx.Done():
					return // Stop accepting new connections on shutdown
				default:
					continue // Ignore other situations and keep accepting
				}
			}

			balancingWG.Add(1)
			go func() {
				defer balancingWG.Done()
				b.handleConnection(clientConn, connectionsCtx)
			}()
		}
	}()

	// Wait for shutdown signal
	select {
	// Graceful shutdowns allow existing connections to finish
	case <-gracefulShutdownCtx.Done():
		listener.Close() // Stop accepting new connections
	case <-sigTermCtx.Done():
		listener.Close() // Stop accepting new connections
	// Immediate shutdowns forcefully close all connections
	case <-immediateShutdownCtx.Done():
		listener.Close() // Stop accepting new connections
		cancel()         // Cancel the connections context to signal all handlers to stop immediately
	}

	// Wait for all ongoing connections to finish
	balancingWG.Wait()
	cancel() // Finally cancel the context as a safety measure
}

func (b *Balancer) handleConnection(clientConn net.Conn, stopCtx context.Context) {
	backend, err := b.Strategy.PickBackend()
	if err != nil {
		clientConn.Close() // Close client connection if no backend is available
		return
	}

	backendConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", backend.Address, backend.Port))
	if err != nil {
		clientConn.Close() // Close client connection if backend connection fails
		return
	}

	connectionWG := &sync.WaitGroup{}
	connectionWG.Add(2)

	// Client -> Backend
	go func() {
		defer connectionWG.Done()
		io.Copy(backendConn, clientConn)
		backendConn.(*net.TCPConn).CloseWrite() // Signal second goroutine we're done sending data
	}()
	// Backend -> Client
	go func() {
		defer connectionWG.Done()
		io.Copy(clientConn, backendConn)
		clientConn.(*net.TCPConn).CloseWrite() // Signal first goroutine we're done sending data
	}()

	// Connection wait group wrapper
	done := make(chan struct{})
	go func() {
		connectionWG.Wait()
		close(done)
	}()

	// Wait for both directions to finish or for a shutdown signal
	select {
	case <-done:
		// Close connections explicitly
		clientConn.Close()
		backendConn.Close()
		return // Both directions finished
	case <-stopCtx.Done():
		clientConn.Close()
		backendConn.Close()
		connectionWG.Wait() // Wait for ongoing transfers to finish
		return
	}
}

/* Managing configuration */

func (b *Balancer) UpdateLoadBalancingStrategy(strategy string) {
	// TODO: Implement this method to update the load balancing strategy
}

func (b *Balancer) UpdateConnectionTimeoutSec(timeout int) {
	b.ConnectionTimeoutSecMu.Lock()
	defer b.ConnectionTimeoutSecMu.Unlock()
	b.Config.ConnectionTimeoutSec = timeout
}

func (b *Balancer) UpdateMaxConnections(max int) {
	b.MaxConnectionsMu.Lock()
	defer b.MaxConnectionsMu.Unlock()
	b.Config.MaxConnections = max
}

func (b *Balancer) UpdateBackends(backends []string) {
	// TODO: Implement this method to update the list of backends
}

/* Monitoring health and performance */
