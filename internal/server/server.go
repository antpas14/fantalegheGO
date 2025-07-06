package server

import (
	"fmt"
	"net/http"

	api "github.com/antpas14/fantalegheEV-api" // Alias as 'api' for cleaner usage
	echo "github.com/labstack/echo/v4"

	"fantalegheGO/internal/calculate"
	"fantalegheGO/internal/excel"
	"fantalegheGO/internal/parser"
)

type MyServer struct {
	e                *echo.Echo
	calculateService calculate.Calculate
}

func NewMyServer() *MyServer {
	e := echo.New()

	parserInstance := parser.NewParserImpl()
	excelServiceInstance := excel.NewExcelService()
	calculateServiceInstance := calculate.NewCalculateImpl(excelServiceInstance, parserInstance)

	server := &MyServer{
		e:                e,
		calculateService: calculateServiceInstance,
	}
	server.setupRoutes()

	return server
}

func (s *MyServer) setupRoutes() {
	api.RegisterHandlers(s.e, s)
}

func (s *MyServer) Serve(port string) error {
	fmt.Printf("Starting HTTP server on port %s...\n", port)
	// Echo's Start method is blocking, it will keep the server running until stopped.
	return s.e.Start(port)
}

func (s *MyServer) Calculate(ctx echo.Context) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
	}

	files := form.File["file"]
	if len(files) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No file uploaded. Please provide an Excel file.")
	}

	uploadedFileHeader := files[0]

	fmt.Printf("Uploaded File: %s, Size: %d bytes\n", uploadedFileHeader.Filename, uploadedFileHeader.Size)

	ranks, err := s.calculateService.GetRanks(uploadedFileHeader)
	if err != nil {
		ctx.Logger().Errorf("Error during calculation for file '%s': %v", uploadedFileHeader.Filename, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Calculation failed: "+err.Error())
	}
	return ctx.JSON(http.StatusOK, ranks)
}
