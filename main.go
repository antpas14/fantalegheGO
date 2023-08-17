package main

import (
	"github.com/labstack/echo/v4"
	"github.com/antpas14/fantalegheEV-api"
	"fantalegheGO/internal/server"
)

func main() {
	e := echo.New()

	// Create an instance of your server implementation
	serverImpl := &server.MyServer{}

	// Register the server handlers
	api.RegisterHandlers(e, serverImpl)

	// Start the server
	e.Start(":8080")
}