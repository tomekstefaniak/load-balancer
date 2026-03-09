package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"load-balancer/internal/balancer"
	cmn "load-balancer/internal/common"
	"load-balancer/internal/config"
	"load-balancer/internal/strategy"
)

type UI struct {
	Config          *config.Config
	Balancer        *balancer.Balancer
	GracefulCtx     context.Context
	GracefulCancel  context.CancelFunc
	ImmediateCtx    context.Context
	ImmediateCancel context.CancelFunc
	StateModMu      sync.Mutex // Mutex to protect state modifications across handlers
}

func (u *UI) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", u.handleStatus)
	mux.HandleFunc("/strategy", u.handleStrategy)
	mux.HandleFunc("/backend/add", u.handleBackendAdd)
	mux.HandleFunc("/backend/remove", u.handleBackendRemove)
	mux.HandleFunc("/backend/ls", u.handleBackendLs)
	mux.HandleFunc("/backend/conns", u.handleBackendConns)
	mux.HandleFunc("/timeout", u.handleTimeout)
	mux.HandleFunc("/max-connections", u.handleMaxConnections)
	mux.HandleFunc("/stop", u.handleStop)
	mux.HandleFunc("/start", u.handleStart)
	mux.HandleFunc("/shutdown", u.handleShutdown)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", u.Config.ClientPort),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("management server error: %v\n", err)
	}
}

// GET /status
func (u *UI) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	b := u.Balancer

	b.StateMu.RLock()
	state := b.State
	b.StateMu.RUnlock()

	b.LoadBalancingStrategyMu.RLock()
	strategy := u.Config.LoadBalancingStrategy
	b.LoadBalancingStrategyMu.RUnlock()

	b.CurrConnectionsMu.Lock()
	currConns := b.CurrConnections
	b.CurrConnectionsMu.Unlock()

	b.MaxConnectionsMu.Lock()
	maxConns := u.Config.MaxConnections
	b.MaxConnectionsMu.Unlock()

	b.IdleTimeoutSecMu.RLock()
	idleTimeout := u.Config.IdleTimeoutSec
	b.IdleTimeoutSecMu.RUnlock()

	b.BackendsMu.RLock()
	backends := make([]cmn.Backend, len(u.Config.Backends))
	copy(backends, u.Config.Backends)
	b.BackendsMu.RUnlock()

	stateNames := map[int]string{
		balancer.BUSY:                 "busy",
		balancer.IDLE:                 "idle",
		balancer.STOPPING_GRACEFULLY:  "stopping_gracefully",
		balancer.STOPPING_IMMEDIATELY: "stopping_immediately",
	}

	resp := map[string]any{
		"state":               stateNames[state],
		"strategy":            config.StrategyName(strategy),
		"current_connections": currConns,
		"max_connections":     maxConns,
		"idle_timeout_sec":    idleTimeout,
		"listener_port":       u.Config.ListenerPort,
		"backends":            backends,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /strategy {"strategy": "roundrobin"}
func (u *UI) handleStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Strategy string `json:"strategy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	strategyType, ok := config.ParseStrategy(req.Strategy)
	if !ok {
		http.Error(w, `{"error":"unknown strategy, use: roundrobin, leastconnections, random"}`, http.StatusBadRequest)
		return
	}

	u.Balancer.UpdateLoadBalancingStrategy(strategyType)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"strategy":"%s"}`, config.StrategyName(strategyType))
}

// POST /backend/add {"address":"127.0.0.1", "port":8083, "max_connections":100}
func (u *UI) handleBackendAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address        string `json:"address"`
		Port           int    `json:"port"`
		MaxConnections int    `json:"max_connections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, `{"error":"address is required"}`, http.StatusBadRequest)
		return
	}
	if req.Port <= 0 || req.Port > 65535 {
		http.Error(w, `{"error":"port must be between 1 and 65535"}`, http.StatusBadRequest)
		return
	}
	if req.MaxConnections <= 0 {
		http.Error(w, `{"error":"max_connections must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	u.Balancer.AddBackend(cmn.Backend{
		Address:        req.Address,
		Port:           req.Port,
		MaxConnections: req.MaxConnections,
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}

// POST /backend/remove {"address":"127.0.0.1", "port":8083}
func (u *UI) handleBackendRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, `{"error":"address is required"}`, http.StatusBadRequest)
		return
	}
	if req.Port <= 0 || req.Port > 65535 {
		http.Error(w, `{"error":"port must be between 1 and 65535"}`, http.StatusBadRequest)
		return
	}

	u.Balancer.RemoveBackend(req.Address, req.Port)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}

// GET /backend/ls
func (u *UI) handleBackendLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	u.Balancer.BackendsMu.RLock()
	backends := make([]cmn.Backend, len(u.Config.Backends))
	copy(backends, u.Config.Backends)
	u.Balancer.BackendsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"backends": backends,
		"count":    len(backends),
	})
}

// GET /backend/conns
func (u *UI) handleBackendConns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	u.Balancer.BackendConnsMu.Lock()
	conns := make(map[string]int, len(u.Balancer.BackendConns.Conns))
	for k, v := range u.Balancer.BackendConns.Conns {
		conns[k] = v
	}
	u.Balancer.BackendConnsMu.Unlock()

	// Build response with backend info alongside connection counts
	u.Balancer.BackendsMu.RLock()
	type backendConns struct {
		Address        string `json:"address"`
		Port           int    `json:"port"`
		Connections    int    `json:"connections"`
		MaxConnections int    `json:"max_connections"`
	}
	result := make([]backendConns, 0, len(u.Config.Backends))
	for _, b := range u.Config.Backends {
		result = append(result, backendConns{
			Address:        b.Address,
			Port:           b.Port,
			Connections:    conns[strategy.BackendKey(b)],
			MaxConnections: b.MaxConnections,
		})
	}
	u.Balancer.BackendsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// POST /timeout {"idle_timeout_sec": 60}
func (u *UI) handleTimeout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IdleTimeoutSec int `json:"idle_timeout_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.IdleTimeoutSec <= 0 {
		http.Error(w, `{"error":"idle_timeout_sec must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	u.Balancer.UpdateIdleTimeoutSec(req.IdleTimeoutSec)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"idle_timeout_sec":%d}`, req.IdleTimeoutSec)
}

// POST /max-connections {"max_connections": 500}
func (u *UI) handleMaxConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MaxConnections int `json:"max_connections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.MaxConnections <= 0 {
		http.Error(w, `{"error":"max_connections must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	u.Balancer.UpdateMaxConnections(req.MaxConnections)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"max_connections":%d}`, req.MaxConnections)
}

// POST /stop {"mode": "graceful"} or {"mode": "immediate"}
func (u *UI) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Lock to prevent concurrent stop/start operations
	u.StateModMu.Lock()
	defer u.StateModMu.Unlock()

	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Mode {
	case "graceful":
		fmt.Fprintf(w, `{"ok":true,"mode":"graceful"}`)
		go u.GracefulCancel()
	case "immediate":
		fmt.Fprintf(w, `{"ok":true,"mode":"immediate"}`)
		go u.ImmediateCancel()
	default:
		http.Error(w, `{"error":"mode must be 'graceful' or 'immediate'"}`, http.StatusBadRequest)
	}
}

// POST /start {}
func (u *UI) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Lock to prevent concurrent stop/start operations
	u.StateModMu.Lock()
	defer u.StateModMu.Unlock()

	u.Balancer.StateMu.RLock()
	if u.Balancer.State != balancer.IDLE {
		u.Balancer.StateMu.RUnlock()
		http.Error(w, `{"error":"balancer is not idle so it cannot be started"}`, http.StatusConflict)
		return
	}
	u.Balancer.StateMu.RUnlock()

	// Create new contexts for the balancer and start it
	u.GracefulCtx, u.GracefulCancel = context.WithCancel(context.Background())
	u.ImmediateCtx, u.ImmediateCancel = context.WithCancel(context.Background())
	go u.Balancer.Start(u.GracefulCtx, u.ImmediateCtx)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"message":"balancer started"}`)
}

// POST /shutdown {"timeout_sec": 5}
func (u *UI) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Lock to prevent concurrent stop/start operations
	u.StateModMu.Lock()
	defer u.StateModMu.Unlock()

	var req struct {
		TimeoutSec int `json:"timeout_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 5
	}

	// Force immediate stop of the balancer
	u.ImmediateCancel()

	// Wait for balancer to reach IDLE state, up to timeout
	deadline := time.After(time.Duration(req.TimeoutSec) * time.Second)
	for {
		u.Balancer.StateMu.RLock()
		idle := u.Balancer.State == balancer.IDLE
		u.Balancer.StateMu.RUnlock()
		if idle {
			break
		}
		select {
		case <-deadline:
			os.Exit(1) // Exit with error if balancer did not shut down in time
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	os.Exit(0) // Finally exit with 0 once balancer successfully shuts down
}
