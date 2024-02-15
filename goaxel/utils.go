package goaxel

import (
	"errors"
	"fmt"
	"math"
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

func printProgress(progress <-chan uint64, totalContentBytes uint64) {
	progressBytes := uint64(0)
	startTime := time.Now().UnixMilli()
	speedUnits := []string{"B/s ", "KB/s", "MB/s", "GB/s", "TB/s"}

	for workerProgress, ok := <-progress; ok; workerProgress, ok = <-progress {
		progressBytes += workerProgress

		currentTime := time.Now().UnixMilli()
		duration := float64(currentTime-startTime) / 1000.0 // Seconds

		progressPercent := float64(progressBytes) * 100.0 / float64(totalContentBytes)

		speed := float64(progressBytes) / duration
		speedUnit := speedUnits[0]
		for i := 1; i < len(speedUnits); i++ {
			if speed > 1024 {
				speed /= 1024
				speedUnit = speedUnits[i]
			} else {
				break
			}
		}
		if speed > 1024 {
			speed = math.Max(speed, 9999.99)
			speedUnit = "fast"
		}

		fmt.Printf("\rprogress: %6.2f%% speed: %7.2f%v", progressPercent, speed, speedUnit)
	}
}
