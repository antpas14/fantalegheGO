package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	api "github.com/antpas14/fantalegheEV-api"
	"github.com/labstack/echo/v4"
)

type MockCalculate struct {
	GetRanksFunc func(fileHeader *multipart.FileHeader) ([]api.Rank, error)
}

func (m *MockCalculate) GetRanks(fileHeader *multipart.FileHeader) ([]api.Rank, error) {
	if m.GetRanksFunc != nil {
		return m.GetRanksFunc(fileHeader)
	}
	return nil, errors.New("GetRanks not implemented in MockCalculate")
}

type MockFileHeaderOpener struct {
	OpenFunc func() (io.Reader, error)
	FileName string
	FileSize int64
}

func (m *MockFileHeaderOpener) Open() (io.Reader, error) { return m.OpenFunc() }
func (m *MockFileHeaderOpener) Filename() string         { return m.FileName }
func (m *MockFileHeaderOpener) Size() int64              { return m.FileSize }

// --- TestMyServer Tests ---

func TestCalculateEndpoint(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name               string
		mockCalculate      *MockCalculate
		fileContent        string
		fileName           string
		fileMimeType       string
		expectStatusCode   int
		expectBodyContains string
		expectRanks        []api.Rank
	}{
		{
			name: "Successful calculation",
			mockCalculate: &MockCalculate{
				GetRanksFunc: func(fileHeader *multipart.FileHeader) ([]api.Rank, error) {
					return []api.Rank{
						{Team: apiString("TeamA"), Points: apiInt(10), EvPoints: apiFloat64(5.5)},
					}, nil
				},
			},
			fileContent:      "dummy excel data",
			fileName:         "test.xlsx",
			fileMimeType:     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			expectStatusCode: http.StatusOK,
			expectRanks: []api.Rank{
				{Team: apiString("TeamA"), Points: apiInt(10), EvPoints: apiFloat64(5.5)},
			},
		},
		{
			name: "No file uploaded",
			mockCalculate: &MockCalculate{
				GetRanksFunc: func(fileHeader *multipart.FileHeader) ([]api.Rank, error) {
					return nil, nil
				},
			},
			fileContent:        "",
			fileName:           "",
			expectStatusCode:   http.StatusBadRequest,
			expectBodyContains: "No file uploaded",
		},
		{
			name: "CalculateService returns error",
			mockCalculate: &MockCalculate{
				GetRanksFunc: func(fileHeader *multipart.FileHeader) ([]api.Rank, error) {
					return nil, errors.New("internal calculation error")
				},
			},
			fileContent:        "some excel data",
			fileName:           "error.xlsx",
			fileMimeType:       "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			expectStatusCode:   http.StatusInternalServerError,
			expectBodyContains: "Calculation failed: internal calculation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &MyServer{
				e:                e,
				calculateService: tt.mockCalculate,
			}
			server.setupRoutes()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.fileName != "" {
				part, err := writer.CreateFormFile("file", tt.fileName)
				if err != nil {
					t.Fatalf("Failed to create form file: %v", err)
				}
				_, err = io.Copy(part, bytes.NewReader([]byte(tt.fileContent)))
				if err != nil {
					t.Fatalf("Failed to write file content: %v", err)
				}
			}
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/calculate", body)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectStatusCode {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectStatusCode, rec.Code, rec.Body.String())
			}

			if tt.expectBodyContains != "" {
				if !strings.Contains(rec.Body.String(), tt.expectBodyContains) {
					t.Errorf("Expected body to contain '%s', got '%s'", tt.expectBodyContains, rec.Body.String())
				}
			}

			if tt.expectStatusCode == http.StatusOK && tt.expectRanks != nil {
				var gotRanks []api.Rank
				err := json.Unmarshal(rec.Body.Bytes(), &gotRanks)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if !reflect.DeepEqual(gotRanks, tt.expectRanks) {
					t.Errorf("Expected ranks %v, got %v", tt.expectRanks, gotRanks)
				}
			}
		})
	}
}

// Helper functions

func apiString(s string) *string {
	return &s
}

func apiInt(i int) *int {
	return &i
}

func apiFloat64(f float64) *float64 {
	return &f
}
