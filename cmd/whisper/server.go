package main

import (
	"log"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/mutablelogic/go-whisper/pkg/whisper/api"
)

type ServerCmd struct {
	Endpoint string `name:"endpoint" help:"Endpoint for the server" default:"/api/v1"`
	Listen   string `name:"listen" help:"Listen address for the server" default:"localhost:8080"`
}

func (cmd *ServerCmd) Run(ctx *Globals) error {
	// Create a mux for serving requests, then register the endpoints with the mux
	mux := http.NewServeMux()

	// Register the endpoints
	api.RegisterEndpoints(cmd.Endpoint, mux, ctx.service)

	// Create a new HTTP server
	log.Println("List address", cmd.Listen)
	server, err := httpserver.Config{
		Listen: cmd.Listen,
		Router: mux,
	}.New()
	if err != nil {
		return err
	}

	// Run the server until CTRL+C
	log.Println("Press CTRL+C to exit")
	return server.Run(ctx.ctx)
}
