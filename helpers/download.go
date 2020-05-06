package helpers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Download -- Download from a given url
func Download(directory string, filename string, url string) error {
	fullFilePath := filepath.Join(directory, filename)
	info, err := os.Stat(fullFilePath)
	if info != nil {
		return fmt.Errorf("%s already exists", info.Name())
	}
	fmt.Printf("Downloading %s \n", filename)

	mode := int(0755)
	err = os.MkdirAll(directory, os.FileMode(mode))
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return err
}
