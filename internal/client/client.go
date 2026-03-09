package client

import (
	"context"

	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
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
