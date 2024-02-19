package goaxel

import (
	"fmt"
	"os"
	"strings"
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

	// check if metadata file exists
	metadataFilename := ""
	metadata := NewMetadata(conn)
	listDirectories, err := os.ReadDir("./")
	if err != nil {
		fmt.Println("Error: Couldnot list the current directory to check goaxel metadata file")
		os.Exit(1)
	}
	for _, entry := range listDirectories {
		if entry.Type().IsRegular() &&
			strings.HasPrefix(entry.Name(), downloadInfo.Filename) &&
			strings.HasSuffix(entry.Name(), ".gst") {
			metadataFilename = entry.Name()
			break
		}
	}

	if metadataFilename == "" {
		// create metadata file
		fmt.Println("No Metadata file found")
		for i := uint64(0); i < conn; i++ {
			metadata.ranges[i].start = i * (downloadInfo.ContentLength / conn)
			metadata.ranges[i].current = metadata.ranges[i].start
			metadata.ranges[i].stop = metadata.ranges[i].start + (downloadInfo.ContentLength / conn) - 1
			if i == conn-1 {
				metadata.ranges[i].stop = downloadInfo.ContentLength - 1
			}
		}
		metadataFilename = downloadInfo.Filename + ".gst"
		if err := CreateMetadata(metadataFilename); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
		if err := WriteMetadata(metadataFilename, metadata); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
	} else { // if metadata already exists
		fmt.Println("Metadata file found")
		metadata, err = ReadMetadata(metadataFilename)
		if err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
	}

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
		printProgress(progress, metadataFilename)
	}()

	for i := uint64(0); i < conn; i++ {
		workerWg.Add(1)
		go func(workerId, start, stop, current uint64) {
			defer workerWg.Done()
			DownloadRange(workerId, progress, downloadInfo, downloadInfo.Filename, buffer_size, start, stop, current)
		}(i, metadata.ranges[i].start, metadata.ranges[i].stop, metadata.ranges[i].current)
	}

	workerWg.Wait()
	close(progress)
	progressWg.Wait()

	// destroy Metadata file
	if err := DestroyMetadata(metadataFilename); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}
