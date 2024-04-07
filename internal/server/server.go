package server

import (
	"net/http"

	"fantalegheGO/internal/calculate"
	"fantalegheGO/internal/parser"
	api "github.com/antpas14/fantalegheEV-api"
	echo "github.com/labstack/echo/v4"
)

// Define your server implementation
type MyServer struct{}

var calculateInstance = calculate.CalculateImpl{}
var parserInstance parser.Parser = parser.DefaultParserImpl()

func (s *MyServer) Calculate(ctx echo.Context, params api.CalculateParams) error {
	if params.LeagueName == nil {
		// Handle the case where LeagueName is nil, perhaps return an error.
		return ctx.JSON(http.StatusBadRequest, "LeagueName is required")
	}

	// Get ranks from calculate module
	ranks, _ := calculateInstance.GetRanks(*params.LeagueName, parserInstance)

	return ctx.JSON(http.StatusOK, ranks)
}
