package excel

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

type ExcelServiceImpl struct{}

func NewExcelService() *ExcelServiceImpl {
	return &ExcelServiceImpl{}
}

type FileHeaderOpener interface {
	Open() (multipart.File, error) // <-- THIS IS THE CRITICAL SIGNATURE
}

func (es *ExcelServiceImpl) ReadExcelFromReader(reader io.Reader) ([][]string, error) {
	if reader == nil {
		return nil, fmt.Errorf("excel: reader is nil")
	}

	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("excel: failed to open Excel file with excelize: %w", err)
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel: no sheets found in the Excel file")
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("excel: failed to get rows from sheet '%s': %w", sheetName, err)
	}

	var filteredRows [][]string
	for _, row := range rows {
		isEmpty := true

		for _, cell := range row {
			if cell != "" {
				isEmpty = false
				break
			}
		}
		if !isEmpty {
			// Row is not empty, but we also want to remove all empty cells from row
			var newRow []string
			for _, str := range row {
				if str != "" {
					newRow = append(newRow, str)
				}
			}

			filteredRows = append(filteredRows, newRow)
		}
	}

	if len(filteredRows) == 0 {
		return nil, fmt.Errorf("excel: file contains no valid data rows after filtering")
	}

	return filteredRows, nil
}

func (es *ExcelServiceImpl) ReadExcel(fileHeader FileHeaderOpener) ([][]string, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("file header is nil")
	}

	file, err := fileHeader.Open() // This call expects the multipart.File, error signature
	if err != nil {
		return nil, fmt.Errorf("excel: failed to open uploaded file: %w", err)
	}
	defer file.Close()

	return es.ReadExcelFromReader(file)
}
