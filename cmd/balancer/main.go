package main

import (
	"context"

	"load-balancer/internal/balancer"
	"load-balancer/internal/client"
	"load-balancer/internal/config"
	"load-balancer/internal/flags"
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
	balancer.Balance(gracefulCtx, immediateCtx)

	// Start client listener in a separate goroutine
	client := &client.Client{
		Config:   cfg,
		Balancer: balancer,
	}

	client.Start(gracefulCtx, immediateCtx) // Start the client listener
}
