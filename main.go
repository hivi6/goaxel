package main

import (
  "fmt"
  "net/http"
  "os"
  "io"
  "strconv"
  "time"
)

func main() {
  if len(os.Args) <= 1 {
    fmt.Println("Usage: goaxel <url>")
    os.Exit(1)
  }
  url := os.Args[1]

  // create a client for the request
  client := &http.Client {}

  // creating a request
  req, err := http.NewRequest("HEAD", url, nil)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while generating the HEAD request")
    os.Exit(1)
  }
  fmt.Println("method:", req.Method)

  // making the request and getting the response
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while making the request")
    os.Exit(1)
  }
  fmt.Println("status code:", resp.StatusCode)

  // checking if Accept-Ranges key is available
  acceptRanges := resp.Header.Get("Accept-Ranges")
  fmt.Println("Accept-Ranges:", acceptRanges)
  if acceptRanges == "" || acceptRanges == "none" {
    fmt.Println("Error: Server doesn't accepts ranges")
    os.Exit(1)
  }

  // getting the Content-Length
  contentLengthStr := resp.Header.Get("Content-Length")
  fmt.Println("Content-Length:", contentLengthStr)
  if contentLengthStr == "" {
    fmt.Println("Error: Server didn't provide any contentLength")
    os.Exit(1)
  }

  // Converting the Content-Length to Integer
  contentLength, err := strconv.Atoi(contentLengthStr)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while converting Content-Length to Integer")
    os.Exit(1)
  }

  // Creating a GET request
  req, err = http.NewRequest("GET", url, nil)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while generating the GET request")
    os.Exit(1)
  }
  // add the range
  req.Header.Add("Range", fmt.Sprintf("bytes=0-%v", contentLength))
  fmt.Println("method:", req.Method)

  // making the request and getting the response
  resp, err = client.Do(req)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while making the request")
    os.Exit(1)
  }
  fmt.Println("status code:", resp.StatusCode)
  defer resp.Body.Close()

  // create a file a read to that file
  file, _ := os.Create("temp.jpeg")
  remainingLength := contentLength
  startTime := time.Now()
  fmt.Println("startTime:", startTime)
  for remainingLength > 0 {
    fmt.Printf("\r")
    bufferSize := 10 * 1024
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
    fmt.Printf("speed: %vKBps   ", float64(contentLength - remainingLength) / diffTimeSeconds / 1024.0)
  }
  fmt.Println()
}
