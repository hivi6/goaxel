package goaxel

import (
	"fmt"
	"math"
	"time"
)

type ProgressInfo struct {
	workerId uint64
	start    uint64
	stop     uint64
	current  uint64
}

func printProgress(progress <-chan ProgressInfo, connections, totalContentBytes uint64) {
	progressBytes := uint64(0)
	startTime := time.Now().UnixMilli()
	speedUnits := []string{"B/s ", "KB/s", "MB/s", "GB/s", "TB/s"}
	workerProgresses := make([]ProgressInfo, connections)

	for workerProgress, ok := <-progress; ok; workerProgress, ok = <-progress {
		workerProgresses[workerProgress.workerId] = workerProgress

		progressBytes = 0
		for _, workerProgress := range workerProgresses {
			progressBytes += workerProgress.current - workerProgress.start
		}

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

	currentTime := time.Now().UnixMilli()
	fmt.Printf("\n\nElapsed-Time: %.2fs\n", float64(currentTime-startTime)/1000.0)
}
