package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/antpas14/fantalegheEV-api"
	"fantalegheGO/internal/calculate"
	"fantalegheGO/internal/parser"

)

// Define your server implementation
type MyServer struct{}

var parserInstance = parser.ParserImpl{}

func (s *MyServer) Calculate(ctx echo.Context, params api.CalculateParams) error {
    if params.LeagueName == nil {
		// Handle the case where LeagueName is nil, perhaps return an error.
		return ctx.JSON(http.StatusBadRequest, "LeagueName is required")
	}

	// Get ranks from calculate module
	ranks := calculate.GetRanks(*params.LeagueName, &parserInstance)

	return ctx.JSON(http.StatusOK, ranks)
}
