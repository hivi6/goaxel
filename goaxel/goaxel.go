package goaxel

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func Download(conn uint64, buffer_size uint64, url string) {
	// fetch download information using the url
	downloadInfo, err := FetchDownloadInfo(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Provided Url: %v\n", url)
	fmt.Printf("Number of Connection: %v\n", conn)
	fmt.Printf("Buffer Size: %vKB\n", buffer_size)
	fmt.Printf("Final Url: %v\n", downloadInfo.Url)
	fmt.Printf("Content Size: %vB\n", downloadInfo.ContentLength)

	// create a file a write to that file
	finalFilename, file, err := CreateFile(downloadInfo.Filename)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	defer file.Close()
	// create a file with size as the contentLength
	_, err = file.Seek(int64(downloadInfo.ContentLength-1), 0)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Couldnot create file with given size")
		os.Exit(1)
	}
	_, err = file.Write([]byte{0})
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Couldnot create file with given size")
		os.Exit(1)
	}

	startTime := time.Now()

	progress := make(chan uint64)
	var progressWg sync.WaitGroup
	var workerWg sync.WaitGroup

	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		printProgress(progress, downloadInfo.ContentLength)
	}()

	numberOfWorker := conn
	for i := uint64(0); i < numberOfWorker; i++ {
		start := i * (downloadInfo.ContentLength / numberOfWorker)
		stop := start + (downloadInfo.ContentLength / numberOfWorker) - 1
		if i == numberOfWorker-1 {
			stop = downloadInfo.ContentLength - 1
		}
		workerWg.Add(1)
		go func(start, stop uint64) {
			defer workerWg.Done()
			DownloadRange(progress, downloadInfo, finalFilename, buffer_size, start, stop)
		}(start, stop)
	}

	workerWg.Wait()
	close(progress)
	progressWg.Wait()

	currentTime := time.Now()
	diffTime := currentTime.Sub(startTime)
	diffSeconds := diffTime.Seconds()
	fmt.Printf("\n\nElapsed-Time: %.2fs\n", diffSeconds)
}
