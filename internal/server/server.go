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
	// Get ranks from the data package
	ranks := calculate.GetRanks()

	return ctx.JSON(http.StatusOK, ranks)
}
// Helper function to create a float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create an int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}
