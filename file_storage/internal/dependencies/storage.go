package dependencies

import (
	"file-storage/internal/general"
	"io"
)

type StorageHandler interface {
	SaveToFileB(filePart []byte, location string) (int64, *general.ErrorResult)
	SaveToFile(filePart io.Reader, location string) (int64, *general.ErrorResult)
	GetDataFolder(name string) (string, error)
}
