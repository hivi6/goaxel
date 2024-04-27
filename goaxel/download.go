package goaxel

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"os"
	"strconv"
	"strings"
)

type DownloadInfo struct {
	Url string

	ContentLength uint64
	AcceptRanges  bool

	Filename string
}

func FetchDownloadInfo(url string) (DownloadInfo, error) {
	// create a client for the request
	client, err := CreateClient()
	if err != nil {
		return DownloadInfo{}, errors.New("something went wrong while creating HTTP Client")
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return DownloadInfo{}, errors.New("something went wrong while creating the HEAD request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return DownloadInfo{}, errors.New("something went wrong while making the HEAD request")
	}
	defer resp.Body.Close()

	contentLengthStr := resp.Header.Get("Content-Length")
	if contentLengthStr == "" {
		return DownloadInfo{}, errors.New("no Content-Length was provided in the response by the server")
	}
	contentLength, err := strconv.ParseUint(contentLengthStr, 10, 64)
	if err != nil {
		return DownloadInfo{}, errors.New("something went wrong while parsing Content-Length to uint64")
	}

	acceptRangesStr := resp.Header.Get("Accept-Ranges")
	acceptRanges := (acceptRangesStr == "bytes")
	if acceptRangesStr == "" {
		acceptRanges = false
	} else if acceptRangesStr != "none" && acceptRangesStr != "bytes" {
		msg := fmt.Sprintf("Accept-Ranges of type '%s' not supported", acceptRangesStr)
		return DownloadInfo{}, errors.New(msg)
	}

	// generating outputfilename
	outputFilename := ""
	// TODO: try to get filename from Content-Disposition
	// for now generating filename using url
	if outputFilename == "" {
		parsedUrl, err := netUrl.Parse(url)
		if err != nil {
			return DownloadInfo{}, errors.New("something went wrong while parsing url")
		}
		if parsedUrl.Path != "" {
			urlPathSplited := strings.Split(parsedUrl.Path, "/")
			outputFilename = urlPathSplited[len(urlPathSplited)-1]
		}
	}
	if outputFilename == "" {
		outputFilename = "default" // this is the default filename
	}

	downloadInfo := DownloadInfo{
		Url:           url,
		ContentLength: contentLength,
		AcceptRanges:  acceptRanges,
		Filename:      outputFilename,
	}
	return downloadInfo, nil
}

func DownloadRange(workerId uint64, progress chan<- ProgressInfo, downloadInfo DownloadInfo, filename string, buffer_size, start, stop, current uint64) {
	// create a client for the request
	client, err := CreateClient()
	if err != nil {
		fmt.Println("Error: Something went wrong while creating HTTP Client")
		os.Exit(1)
	}

	// Creating a GET request
	req, err := http.NewRequest("GET", downloadInfo.Url, nil)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error: Something went wrong while generating the GET request")
		os.Exit(1)
	}
	// add the range
	bytesRange := fmt.Sprintf("bytes=%v-%v", current, stop)
	req.Header.Add("Range", bytesRange)

	// making the request and getting the response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error: Something went wrong while making the request")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// opening the file and going to the specific start point
	fd, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error: Something went wrong while opening file")
		os.Exit(1)
	}
	_, err = fd.Seek(int64(current), 0)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error: Something went wrong while going to start position of the file")
		os.Exit(1)
	}

	// start reading the body of the request and write it to the file
	contentLength := stop - current + 1
	buffer_size = buffer_size * 1024 // in KB
	buffer := make([]byte, buffer_size)
	workerProgress := uint64(0)
	for contentLength > 0 {
		readSize := buffer_size
		if contentLength < buffer_size {
			readSize = contentLength
		}

		io.ReadFull(resp.Body, buffer[:readSize])
		fd.Write(buffer[:readSize])

		contentLength -= readSize
		workerProgress += readSize

		// non blocking sending progress
		select {
		case progress <- ProgressInfo{
			workerId: workerId,
			start:    start,
			stop:     stop,
			current:  current + workerProgress,
		}:
		default:
		}
	}

	// blocking sending progress
	progress <- ProgressInfo{
		workerId: workerId,
		start:    start,
		stop:     stop,
		current:  current + workerProgress,
	}
}
