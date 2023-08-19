package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/antpas14/fantalegheEV-api"
	"fantalegheGO/internal/calculate"

)

// Define your server implementation
type MyServer struct{}

func (s *MyServer) Calculate(ctx echo.Context, params api.CalculateParams) error {

	// Get ranks from calculate module
	ranks := calculate.GetRanks(*params.LeagueName)

	return ctx.JSON(http.StatusOK, ranks)
}
