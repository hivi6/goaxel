package goaxel

import (
	"fmt"
	"os"
	"sync"
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
	if err := CreateContentFile(downloadInfo.Filename, downloadInfo.ContentLength); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	progress := make(chan ProgressInfo, conn*4)
	var progressWg sync.WaitGroup
	var workerWg sync.WaitGroup

	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		printProgress(progress, conn, downloadInfo.ContentLength)
	}()

	numberOfWorker := conn
	for i := uint64(0); i < numberOfWorker; i++ {
		start := i * (downloadInfo.ContentLength / numberOfWorker)
		stop := start + (downloadInfo.ContentLength / numberOfWorker) - 1
		if i == numberOfWorker-1 {
			stop = downloadInfo.ContentLength - 1
		}
		workerWg.Add(1)
		go func(workerId, start, stop uint64) {
			defer workerWg.Done()
			DownloadRange(workerId, progress, downloadInfo, downloadInfo.Filename, buffer_size, start, stop)
		}(i, start, stop)
	}

	workerWg.Wait()
	close(progress)
	progressWg.Wait()
}
