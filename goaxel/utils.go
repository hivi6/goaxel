package goaxel

import (
	"fmt"
	"net/http"
	"os"
)

func CreateClient() (*http.Client, error) {
	return &http.Client{}, nil
}

func CreateContentFile(filename string, contentSize uint64) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("something went wrong wile creating file '%v' with size '%v'", filename, contentSize)
	}
	defer file.Close()
	if _, err := file.Seek(int64(contentSize)-1, 0); err != nil {
		return fmt.Errorf("something went wrong wile creating file '%v' with size '%v'", filename, contentSize)
	}
	if _, err := file.Write([]byte{0}); err != nil {
		return fmt.Errorf("something went wrong wile creating file '%v' with size '%v'", filename, contentSize)
	}
	return nil
}
