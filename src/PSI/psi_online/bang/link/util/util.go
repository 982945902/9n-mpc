package util

import (
	"os"
	"path/filepath"
)

func IsRecover(store_path string) (bool, error) {
	_, err := os.Stat(filepath.Join(store_path, "BanG"))

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func RunOnce(store_path string) (err error) {
	fd, err := os.Create(filepath.Join(store_path, "BanG"))
	if err != nil {
		return
	}

	fd.Write([]byte("BanG"))

	return
}
