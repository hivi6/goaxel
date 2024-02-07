package main

import (
  "fmt"
  "net/http"
  "os"
  "io"
  "strconv"
  "time"
  "errors"
)

type DownloadInfo struct {
  Url string
  ContentLength uint64
  AcceptRanges bool
}

func CreateClient() (*http.Client, error) {
  client := http.Client{}
  return &client, nil
}

func FetchDownloadInfo(client *http.Client, url string) (*DownloadInfo, error) {
  req, err := http.NewRequest("HEAD", url, nil)
  if err != nil {
    return nil, errors.New("Something went wrong while creating the HEAD request")
  }

  resp, err := client.Do(req)
  if err != nil {
    return nil, errors.New("Something went wrong while making the HEAD request")
  }
  defer resp.Body.Close()

  contentLengthStr := resp.Header.Get("Content-Length")
  if contentLengthStr == "" {
    return nil, errors.New("No Content-Length was provided in the response by the server")
  }
  contentLength, err := strconv.ParseUint(contentLengthStr, 10, 64)
  if err != nil {
    return nil, errors.New("Something went wrong while parsing Content-Length to uint64")
  }

  acceptRangesStr := resp.Header.Get("Accept-Ranges")
  if acceptRangesStr == "" {
    return nil, errors.New("No Accept-Ranges was provided in the response by the server")
  }
  if acceptRangesStr != "none" && acceptRangesStr != "bytes" {
    msg := fmt.Sprintf("Accept-Ranges of type '%s' not supported", acceptRangesStr)
    return nil, errors.New(msg)
  }
  acceptRanges := (acceptRangesStr == "bytes")

  downloadInfo := DownloadInfo{
    Url: url,
    ContentLength: contentLength,
    AcceptRanges: acceptRanges,
  }
  return &downloadInfo, nil
}

func main() {
  if len(os.Args) <= 1 {
    fmt.Println("Usage: goaxel <url>")
    os.Exit(1)
  }
  url := os.Args[1]

  // create a client for the request
  client, err := CreateClient()
  if err != nil {
    fmt.Println("Error: Something went wrong while creating HTTP Client")
    os.Exit(1)
  }

  // fetch download information
  downloadInfo, err := FetchDownloadInfo(client, url)
  if err != nil {
    fmt.Println("Error:", err.Error())
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
  req.Header.Add("Range", fmt.Sprintf("bytes=0-%v", downloadInfo.ContentLength - 1))
  fmt.Println("method:", req.Method)

  // making the request and getting the response
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while making the request")
    os.Exit(1)
  }
  fmt.Println("Response status-code:", resp.StatusCode)
  fmt.Println("Response Content-Length:", resp.Header.Get("Content-Length"))
  defer resp.Body.Close()

  // create a file a read to that file
  file, _ := os.Create("temp.jpeg")
  remainingLength := downloadInfo.ContentLength
  startTime := time.Now()
  fmt.Println("startTime:", startTime)
  for remainingLength > 0 {
    bufferSize := uint64(10 * 1024)
    if remainingLength <= bufferSize {
      bufferSize = remainingLength
    }

    buffer := make([]byte, bufferSize)
    io.ReadFull(resp.Body, buffer)
    file.Write(buffer)

    remainingLength -= bufferSize
    presentTime := time.Now()
    diffTime := presentTime.Sub(startTime)
    diffTimeSeconds := diffTime.Seconds()
    fmt.Printf("\rspeed: %.2fKBps   ", float64(downloadInfo.ContentLength - remainingLength) / diffTimeSeconds / 1024.0)
  }
  fmt.Println()
}
