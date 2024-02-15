package goaxel

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func CreateClient() (*http.Client, error) {
	return &http.Client{}, nil
}

func CreateFile(filename string) (string, *os.File, error) {
	for i := 0; i < 1000; i++ { // should only try 1000 times
		finalFilename := filename
		if i != 0 { // generating filename using unixepoch
			finalFilename = fmt.Sprintf("%v.%v", finalFilename, time.Now().Unix())
		}
		f, err := os.OpenFile(finalFilename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			continue
		}
		return finalFilename, f, nil // was able to create the file
	}

	return "", nil, errors.New("couldnot create file event after trying 1000 times, might add a output filename manually")
}
