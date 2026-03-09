package client

import (
	"context"

	"lil-balancer/balancer"
	"lil-balancer/config"
)

type Client struct {
	Config   *config.Config
	Balancer *balancer.Balancer
}

func (c *Client) Start(
	stopCtx context.Context,
	immediateShutdownCtx context.Context,
) {
	
}
