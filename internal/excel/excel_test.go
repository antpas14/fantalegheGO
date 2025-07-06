package excel

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

type mockMultipartFile struct {
	*multipart.FileHeader
	mockOpenFunc func() (multipart.File, error)
}

func (m *mockMultipartFile) Open() (multipart.File, error) {
	if m.mockOpenFunc != nil {
		return m.mockOpenFunc()
	}
	return nil, fmt.Errorf("mockOpenFunc not set for mockMultipartFile")
}

type mockFile struct {
	io.ReadCloser
	io.ReaderAt
	io.Seeker
}

func newMockFile(r io.Reader) multipart.File {
	fileReader, _ := io.ReadAll(r)
	reader := bytes.NewReader(fileReader)
	return &mockFile{
		ReadCloser: io.NopCloser(reader),
		ReaderAt:   reader,
		Seeker:     reader,
	}
}

func TestExcelService_ReadExcelFromReader(t *testing.T) {
	tests := []struct {
		name        string
		sheetName   string
		excelData   [][]string
		reader      io.Reader
		expected    [][]string
		expectedErr string
	}{
		{
			name:      "Valid Excel file with data",
			sheetName: "Sheet1",
			excelData: [][]string{
				{"Header1", "Header2"},
				{"Row1Col1", "Row1Col2"},
				{"Row2Col1", "Row2Col2"},
			},
			expected: [][]string{
				{"Header1", "Header2"},
				{"Row1Col1", "Row1Col2"},
				{"Row2Col1", "Row2Col2"},
			},
			expectedErr: "",
		},
		{
			name:      "Excel file with empty rows",
			sheetName: "Data",
			excelData: [][]string{
				{"Col1", "Col2"},
				{"Row1Val1", "Row1Val2"},
				{},
				{"Row3Val1", "Row3Val2"},
			},
			expected: [][]string{
				{"Col1", "Col2"},
				{"Row1Val1", "Row1Val2"},
				{"Row3Val1", "Row3Val2"},
			},
			expectedErr: "",
		},
		{
			name:      "Excel file with no data rows (only headers)",
			sheetName: "Headers",
			excelData: [][]string{
				{"H1", "H2"},
			},
			expected:    [][]string{{"H1", "H2"}},
			expectedErr: "",
		},
		{
			name:        "Empty Excel file (no sheets)",
			sheetName:   "EmptySheet",
			excelData:   nil,
			expected:    nil,
			expectedErr: "excel: file contains no valid data rows after filtering",
		},
		{
			name:        "Excel file with an empty sheet (has sheet, no data)",
			sheetName:   "EmptyDataSheet",
			excelData:   [][]string{},
			expected:    nil,
			expectedErr: "excel: file contains no valid data rows after filtering",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := ExcelServiceImpl{}

			var finalReader io.Reader
			if tt.reader != nil {
				finalReader = tt.reader
			} else {
				mockExcelFile := excelize.NewFile()

				if tt.name == "Empty Excel file (no sheets)" {
					mockExcelFile.DeleteSheet("Sheet1")
				} else {
					if tt.sheetName != "Sheet1" {
						_, err := mockExcelFile.NewSheet(tt.sheetName)
						require.NoError(t, err, "Failed to create new sheet")
						sheet, _ := mockExcelFile.GetSheetIndex(tt.sheetName)
						mockExcelFile.SetActiveSheet(sheet)
						mockExcelFile.DeleteSheet("Sheet1")
					} else {
						sheet, _ := mockExcelFile.GetSheetIndex("Sheet1")
						mockExcelFile.SetActiveSheet(sheet)
					}

					for rIdx, row := range tt.excelData {
						for cIdx, cellValue := range row {
							cellRef, err := excelize.CoordinatesToCellName(cIdx+1, rIdx+1)
							require.NoError(t, err, fmt.Sprintf("Failed to get cell name for %d,%d", rIdx, cIdx))
							err = mockExcelFile.SetCellValue(tt.sheetName, cellRef, cellValue)
							require.NoError(t, err, fmt.Sprintf("Failed to set cell value at %s for sheet %s", cellRef, tt.sheetName))
						}
					}
				}

				var b bytes.Buffer
				err := mockExcelFile.Write(&b)
				require.NoError(t, err, "Failed to write mock Excel file to buffer")
				finalReader = bytes.NewReader(b.Bytes())
			}

			actualData, actualErr := es.ReadExcelFromReader(finalReader)

			if tt.expectedErr != "" {
				assert.Error(t, actualErr, "Expected an error but got none for test '%s'", tt.name)
				assert.Contains(t, actualErr.Error(), tt.expectedErr, "Error message mismatch for test '%s'", tt.name)
				assert.Nil(t, actualData, "Expected nil data on error for test '%s'", tt.name)
			} else {
				assert.NoError(t, actualErr, "Did not expect an error but got one for test '%s': %v", tt.name, actualErr)
				assert.Equal(t, tt.expected, actualData, "Parsed data does not match expected data for test '%s'", tt.name)
			}
		})
	}
}

func TestExcelService_ReadExcel(t *testing.T) {
	es := ExcelServiceImpl{}

	tests := []struct {
		name        string
		fileHeader  FileHeaderOpener
		expected    [][]string
		expectedErr string
	}{
		{
			name: "Valid multipart file",
			fileHeader: func() FileHeaderOpener {
				f := excelize.NewFile()
				f.SetCellValue("Sheet1", "A1", "Test")
				f.SetCellValue("Sheet1", "B1", "Data")
				var buf bytes.Buffer
				require.NoError(t, f.Write(&buf), "Failed to write dummy Excel file")

				return &mockMultipartFile{
					FileHeader: &multipart.FileHeader{
						Filename: "valid.xlsx",
						Size:     int64(buf.Len()),
					},
					mockOpenFunc: func() (multipart.File, error) { // This now returns multipart.File
						return newMockFile(&buf), nil // Use our newMockFile helper
					},
				}
			}(),
			expected:    [][]string{{"Test", "Data"}},
			expectedErr: "",
		},
		{
			name:        "Nil file header",
			fileHeader:  nil,
			expected:    nil,
			expectedErr: "file header is nil",
		},
		{
			name: "Open() returns error",
			fileHeader: &mockMultipartFile{
				FileHeader: &multipart.FileHeader{
					Filename: "error.xlsx",
					Size:     100,
				},
				mockOpenFunc: func() (multipart.File, error) {
					return nil, assert.AnError
				},
			},
			expected:    nil,
			expectedErr: assert.AnError.Error(),
		},
		{
			name: "Open() returns malformed data",
			fileHeader: func() FileHeaderOpener {
				return &mockMultipartFile{
					FileHeader: &multipart.FileHeader{
						Filename: "malformed_stream.xlsx",
						Size:     100,
					},
					mockOpenFunc: func() (multipart.File, error) {
						return newMockFile(bytes.NewReader([]byte("malformed stream data"))), nil
					},
				}
			}(),
			expected:    nil,
			expectedErr: "failed to open Excel file with excelize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualData, actualErr := es.ReadExcel(tt.fileHeader)

			if tt.expectedErr != "" {
				assert.Error(t, actualErr, "Expected error for test '%s'", tt.name)
				assert.Contains(t, actualErr.Error(), tt.expectedErr, "Error message mismatch for test '%s'", tt.name)
				assert.Nil(t, actualData, "Expected nil data on error for test '%s'", tt.name)
			} else {
				assert.NoError(t, actualErr, "Did not expect error for test '%s': %v", tt.name, actualErr)
				assert.Equal(t, tt.expected, actualData, "Data mismatch for test '%s'", tt.name)
			}
		})
	}
}
