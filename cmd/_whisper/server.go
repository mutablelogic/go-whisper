package main

import (
	"log"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/mutablelogic/go-whisper/pkg/api"
)

type ServerCmd struct {
	Endpoint string `name:"endpoint" help:"Endpoint for the server" default:"/api/v1"`
	Listen   string `name:"listen" help:"Listen address for the server" default:"localhost:8080"`
}

func (cmd *ServerCmd) Run(ctx *Globals) error {
	// Create a new HTTP server
	log.Println("Listen address", cmd.Listen)
	server, err := httpserver.New(cmd.Listen, api.RegisterEndpoints(cmd.Endpoint, ctx.service, nil, ctx.Debug), nil)
	if err != nil {
		return err
	}

	// Run the server until CTRL+C
	log.Println("Press CTRL+C to exit")
	return server.Run(ctx.ctx)
}
