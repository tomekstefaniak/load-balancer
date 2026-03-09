package main

import (
	"context"

	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/internal/flags"
	"load-balancer/internal/ui"
)

func main() {
	flags, err := flags.ParseFlags() // Parse command-line flags and set up contexts
	if err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfig(flags) // Load configuration from file and override with flags
	if err != nil {
		panic(err)
	}

	// Start the load balancing process
	gracefulCtx, cancelGraceful := context.WithCancel(context.Background())
	immediateCtx, cancelImmediate := context.WithCancel(context.Background())
	defer cancelGraceful()
	defer cancelImmediate()

	balancer := balancer.NewBalancer(cfg) // Create a new balancer instance

	// Start balancer
	go balancer.Start(gracefulCtx, immediateCtx)

	// Start management HTTP server
	mgmt := &ui.UI{
		Config:          cfg,
		Balancer:        balancer,
		GracefulCtx:     gracefulCtx,
		GracefulCancel:  cancelGraceful,
		ImmediateCtx:    immediateCtx,
		ImmediateCancel: cancelImmediate,
	}
	mgmt.Start()
}
