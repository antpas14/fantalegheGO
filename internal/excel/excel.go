package excel

import (
	"io"
)

type ExcelService interface {
	ReadExcelFromReader(reader io.Reader) ([][]string, error)
	ReadExcel(fileHeader FileHeaderOpener) ([][]string, error)
}
