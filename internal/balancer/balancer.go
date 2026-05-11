package balancer

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	cmn "load-balancer/internal/common"
	"load-balancer/internal/config"
	"load-balancer/internal/strategy"
)

const (
	BUSY = iota
	IDLE
	STOPPING_GRACEFULLY
	STOPPING_IMMEDIATELY
)

type Balancer struct {
	Config                  *config.Config
	State                   int
	StateMu                 sync.RWMutex
	Strategy                strategy.Strategy
	LoadBalancingStrategyMu sync.RWMutex
	IdleTimeoutSecMu        sync.RWMutex
	CurrConnectionsMu       sync.Mutex
	CurrConnections         int
	MaxConnectionsMu        sync.RWMutex
	BackendsMu              *sync.RWMutex
	BackendConnsMu          sync.Mutex
	BackendConns            *strategy.BackendConnections
}

func NewBalancer(cfg *config.Config) *Balancer {
	var strat strategy.Strategy
	backendsMu := &sync.RWMutex{}
	backendConns := strategy.NewBackendConnections()
	b := &Balancer{
		Config:          cfg,
		State:           IDLE,
		CurrConnections: 0,
		BackendsMu:      backendsMu,
		BackendConns:    backendConns,
	}
	connMu := &b.BackendConnsMu
	switch cfg.LoadBalancingStrategy {
	case strategy.RoundRobin:
		strat = strategy.NewRoundRobin(&cfg.Backends, backendsMu, backendConns, connMu)
	case strategy.LeastConnections:
		strat = strategy.NewLeastConnections(&cfg.Backends, backendsMu, backendConns, connMu)
	case strategy.Random:
		strat = strategy.NewRandom(&cfg.Backends, backendsMu, backendConns, connMu)
	default:
		panic(fmt.Sprintf("invalid load balancing strategy: %d", cfg.LoadBalancingStrategy))
	}
	b.Strategy = strat

	return b
}

/* Load balancing logic */

func (b *Balancer) Start(
	gracefulShutdownCtx context.Context,
	immediateShutdownCtx context.Context,
) {
	// State management
	b.StateMu.Lock()
	if b.State != IDLE {
		b.StateMu.Unlock()
		panic("balancer is cannot be started")
	}
	b.State = BUSY
	b.StateMu.Unlock()
	// Ensure we set state back to IDLE when balancing ends
	defer func() {
		b.StateMu.Lock()
		b.State = IDLE
		b.StateMu.Unlock()
	}()

	// Start listening for incoming client connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", b.Config.ListenerPort))
	if err != nil {
		panic(fmt.Errorf("failed to start listener: %w", err))
	}

	// Set up SIGTERM signal handling for graceful shutdown
	sigTermCtx, sigTermCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer sigTermCancel() // Close the context when balancing ends

	// Wait group for managing active connections during shutdown
	balancingWG := &sync.WaitGroup{}
	// Context for managing active connections during shutdown
	connectionsCtx, cancelConnections := context.WithCancel(context.Background())

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

			// Check max connections before handling the new connection
			b.CurrConnectionsMu.Lock()
			b.MaxConnectionsMu.RLock()
			if b.CurrConnections >= b.Config.MaxConnections {
				b.CurrConnectionsMu.Unlock()
				b.MaxConnectionsMu.RUnlock()
				clientConn.Close() // Reject new connection if max connections reached
				continue
			}
			b.CurrConnections++
			b.CurrConnectionsMu.Unlock()
			b.MaxConnectionsMu.RUnlock()

			// Handle the connection in a new goroutine
			balancingWG.Add(1)
			go func() {
				defer balancingWG.Done()
				b.handleConnection(clientConn, connectionsCtx)
			}()
		}
	}()

	// Wait for shutdown signal
	select {
	// Graceful shutdown allow existing connections to finish
	case <-gracefulShutdownCtx.Done():
		// Change state to stopping gracefully
		b.StateMu.Lock()
		b.State = STOPPING_GRACEFULLY
		b.StateMu.Unlock()

		listener.Close() // Stop accepting new connections

		balancingDone := make(chan struct{})
		go func() {
			balancingWG.Wait() // Wait for all ongoing connections to finish
			close(balancingDone)
		}()

		select {
		case <-balancingDone:
			// All connections finished gracefully
			cancelConnections() // Finally cancel the context as a safety measure
		case <-immediateShutdownCtx.Done():
			// Change state to stopping immediately
			b.StateMu.Lock()
			b.State = STOPPING_IMMEDIATELY
			b.StateMu.Unlock()

			// If an immediate shutdown signal is received during graceful shutdown, force close all connections
			cancelConnections() // Cancel the connections context to signal all handlers to stop immediately
			balancingWG.Wait()  // Wait for all handlers to acknowledge the cancellation
		}

	// SIGTERM signal initiates graceful shutdown
	case <-sigTermCtx.Done():
		// Change state to stopping gracefully
		b.StateMu.Lock()
		b.State = STOPPING_GRACEFULLY
		b.StateMu.Unlock()

		listener.Close() // Stop accepting new connections

		balancingDone := make(chan struct{})
		go func() {
			balancingWG.Wait() // Wait for all ongoing connections to finish
			close(balancingDone)
		}()

		select {
		case <-balancingDone:
			// All connections finished gracefully
			cancelConnections() // Finally cancel the context as a safety measure
		case <-immediateShutdownCtx.Done():
			// If an immediate shutdown signal is received during graceful shutdown, force close all connections
			cancelConnections() // Cancel the connections context to signal all handlers to stop immediately
			balancingWG.Wait()  // Wait for all handlers to acknowledge the cancellation
		}

	// Immediate shutdowns forcefully close all connections
	case <-immediateShutdownCtx.Done():
		// Change state to stopping immediately
		b.StateMu.Lock()
		b.State = STOPPING_IMMEDIATELY
		b.StateMu.Unlock()

		listener.Close()    // Stop accepting new connections
		cancelConnections() // Cancel the connections context to signal all handlers to stop immediately
		balancingWG.Wait()  // Wait for all ongoing connections to finish

	}
}

func (b *Balancer) handleConnection(clientConn net.Conn, stopCtx context.Context) {
	// Ensure we decrement the current connections count when done
	defer func() {
		b.CurrConnectionsMu.Lock()
		if b.CurrConnections > 0 {
			b.CurrConnections--
		}
		b.CurrConnectionsMu.Unlock()
	}()

	// Capture the current strategy under read lock to ensure consistency during selection
	b.LoadBalancingStrategyMu.RLock()
	strat := b.Strategy
	b.LoadBalancingStrategyMu.RUnlock()

	backend, err := strat.PickBackend()
	if err != nil {
		clientConn.Close() // Close client connection if no backend is available
		return
	}
	defer strat.OnRelease(backend) // Ensure backend is released back to the strategy after handling the connection

	connTimeout := time.Duration(b.Config.ServerConnTimeoutSec) * time.Second
	backendConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", backend.Address, backend.Port), connTimeout)
	if err != nil {
		clientConn.Close() // Close client connection if backend connection fails
		return
	}

	// Read idle timeout
	b.IdleTimeoutSecMu.RLock()
	idleTimeout := time.Duration(b.Config.IdleTimeoutSec) * time.Second
	b.IdleTimeoutSecMu.RUnlock()

	// Wait group for managing goroutines managing bidirectional copy between client and backend
	connectionWG := &sync.WaitGroup{}
	connectionWG.Add(2)

	// Client -> Backend
	go func() {
		defer connectionWG.Done()
		io.Copy(backendConn, &idleTimeoutReader{conn: clientConn, peer: backendConn, timeout: idleTimeout})
		backendConn.(*net.TCPConn).CloseWrite() // Signal second goroutine we're done sending data
	}()
	// Backend -> Client
	go func() {
		defer connectionWG.Done()
		io.Copy(clientConn, &idleTimeoutReader{conn: backendConn, peer: clientConn, timeout: idleTimeout})
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

/* Idle timeout for client-backend communication */

type idleTimeoutReader struct {
	conn    net.Conn
	peer    net.Conn
	timeout time.Duration
}

func (r *idleTimeoutReader) Read(p []byte) (int, error) {
	r.conn.SetReadDeadline(time.Now().Add(r.timeout))
	n, err := r.conn.Read(p)
	if n > 0 {
		r.peer.SetReadDeadline(time.Now().Add(r.timeout))
	}
	return n, err
}

/* Managing configuration */

func (b *Balancer) UpdateLoadBalancingStrategy(strategyType int) {
	var strat strategy.Strategy
	switch strategyType {
	case strategy.RoundRobin:
		strat = strategy.NewRoundRobin(&b.Config.Backends, b.BackendsMu, b.BackendConns, &b.BackendConnsMu)
	case strategy.LeastConnections:
		strat = strategy.NewLeastConnections(&b.Config.Backends, b.BackendsMu, b.BackendConns, &b.BackendConnsMu)
	case strategy.Random:
		strat = strategy.NewRandom(&b.Config.Backends, b.BackendsMu, b.BackendConns, &b.BackendConnsMu)
	default:
		return // Invalid strategy type, ignore the update
	}

	b.LoadBalancingStrategyMu.Lock()
	b.Strategy = strat
	b.Config.LoadBalancingStrategy = strategyType
	b.LoadBalancingStrategyMu.Unlock()
}

func (b *Balancer) UpdateIdleTimeoutSec(timeout int) {
	b.IdleTimeoutSecMu.Lock()
	defer b.IdleTimeoutSecMu.Unlock()
	b.Config.IdleTimeoutSec = timeout
}

func (b *Balancer) UpdateMaxConnections(max int) {
	b.MaxConnectionsMu.Lock()
	defer b.MaxConnectionsMu.Unlock()
	b.Config.MaxConnections = max
}

func (b *Balancer) AddBackend(backend cmn.Backend) {
	b.BackendsMu.Lock()
	defer b.BackendsMu.Unlock()

	for _, existing := range b.Config.Backends {
		if existing.Address == backend.Address && existing.Port == backend.Port {
			return // Already exists
		}
	}

	// Add the new backend
	b.Config.Backends = append(b.Config.Backends, backend)
	// Initialize connection count for the new backend in the strategy
	b.BackendConnsMu.Lock()
	b.BackendConns.Conns[strategy.BackendKey(backend)] = 0
	b.BackendConnsMu.Unlock()
}

func (b *Balancer) RemoveBackend(address string, port int) {
	b.BackendsMu.Lock()
	defer b.BackendsMu.Unlock()

	backends := b.Config.Backends
	for i, existing := range backends {
		if existing.Address == address && existing.Port == port {
			// Remove the backend from the configuration
			b.Config.Backends = append(backends[:i], backends[i+1:]...)
			// Remove connection count for the removed backend
			delete(b.BackendConns.Conns, strategy.BackendKey(existing))
			return
		}
	}
}
