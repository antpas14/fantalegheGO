package calculate

import (
	"mime/multipart"

	api "github.com/antpas14/fantalegheEV-api"
)

type Calculate interface {
	GetRanks(fileHeader *multipart.FileHeader) ([]api.Rank, error)
}
