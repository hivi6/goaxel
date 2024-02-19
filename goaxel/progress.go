package goaxel

import (
	"fmt"
	"math"
	"os"
	"time"
)

type ProgressInfo struct {
	workerId uint64
	start    uint64
	stop     uint64
	current  uint64
}

func printProgress(progress <-chan ProgressInfo, metadataFilename string) {
	startTime := time.Now().UnixMilli()
	speedUnits := []string{"B/s ", "KB/s", "MB/s", "GB/s", "TB/s"}

	metadata, err := ReadMetadata(metadataFilename)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	totalContentBytes := uint64(0)
	totalProgressBytes := uint64(0)
	currentProgressBytes := uint64(0)
	for _, rang := range metadata.ranges {
		totalContentBytes += rang.stop - rang.start + 1
		totalProgressBytes += rang.current - rang.start
	}

	for workerProgress, ok := <-progress; ok; workerProgress, ok = <-progress {
		currentTime := time.Now().UnixMilli()
		duration := float64(currentTime-startTime) / 1000.0 // Seconds

		currentProgressBytes += workerProgress.current - metadata.ranges[workerProgress.workerId].current
		totalProgressBytes += workerProgress.current - metadata.ranges[workerProgress.workerId].current
		progressPercent := float64(totalProgressBytes) * 100.0 / float64(totalContentBytes)
		speed := float64(currentProgressBytes) / duration
		remainingTime := float64(totalContentBytes-totalProgressBytes) / speed

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

		fmt.Printf("\rprogress: %6.2f%% | speed: %7.2f%v | remaining: %8.2fs", progressPercent, speed, speedUnit, remainingTime)

		metadata.ranges[workerProgress.workerId] = MetadataRange{workerProgress.start, workerProgress.stop, workerProgress.current}
		if err := WriteMetadata(metadataFilename, metadata); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
	}

	currentTime := time.Now().UnixMilli()
	fmt.Printf("\n\nElapsed-Time: %.2fs\n", float64(currentTime-startTime)/1000.0)
}
