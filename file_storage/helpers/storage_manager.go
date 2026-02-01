package helpers

import (
	"os"
	"path/filepath"
)

func Get_data_folder(name string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectRoot := cwd
	for {
		fileStoragePath := filepath.Join(projectRoot, "file_storage")
		if _, err := os.Stat(fileStoragePath); err == nil {
			break
		}

		// Move up one directory
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			break
		}
		projectRoot = parent
	}

	return filepath.Join(projectRoot, "file_storage", "data", name)
}
