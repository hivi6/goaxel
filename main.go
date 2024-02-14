package main

import (
  "fmt"
  "net/http"
  "os"
  "io"
  "strconv"
  "time"
  "errors"
  "strings"
  netUrl "net/url"
  "sync"
  "flag"
)

var (
  BUFFER_SIZE uint64
  N_CONNECTIONS uint64
)

type DownloadInfo struct {
  Url string

  ContentLength uint64
  AcceptRanges bool

  OutputFilename string
}

func CreateClient() (*http.Client, error) {
  client := http.Client{}
  return &client, nil
}

func FetchDownloadInfo(url string) (*DownloadInfo, error) {
  // create a client for the request
  client, err := CreateClient()
  if err != nil {
    fmt.Println("Error: Something went wrong while creating HTTP Client")
    os.Exit(1)
  }

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

  // generating outputfilename
  outputFilename := ""
  // TODO: try to get filename from Content-Disposition
  // for now generating filename using url
  if outputFilename == "" {
    parsedUrl, err := netUrl.Parse(url)
    if err != nil {
      return nil, errors.New("Something went wrong while parsing url")
    }
    if parsedUrl.Path != "" {
      urlPathSplited := strings.Split(parsedUrl.Path, "/")
      outputFilename = urlPathSplited[len(urlPathSplited) - 1]
    }
  }
  if outputFilename == "" {
    outputFilename = "default" // this is the default filename
  }

  downloadInfo := DownloadInfo{
    Url: url,
    ContentLength: contentLength,
    AcceptRanges: acceptRanges,
    OutputFilename: outputFilename,
  }
  return &downloadInfo, nil
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
    return finalFilename, f, nil; // was able to create the file
  }

  return "", nil, errors.New("Couldnot create file event after trying 1000 times, might add a output filename manually")
}

func DownloadRange(progress chan<- uint64, downloadInfo *DownloadInfo, filename string, start, stop uint64) {
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
  bytesRange := fmt.Sprintf("bytes=%v-%v", start, stop)
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
  _, err = fd.Seek(int64(start), 0)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("Error: Something went wrong while going to start position of the file")
    os.Exit(1)
  }

  // start reading the body of the request and write it to the file
  var buffer_size = uint64(BUFFER_SIZE * 1024) // 8KB
  contentLength := stop - start + 1
  buffer := make([]byte, buffer_size)
  workerProgress := uint64(0)
  for contentLength > 0 {
    readSize := buffer_size
    if contentLength < buffer_size {
      readSize = contentLength
    }

    io.ReadFull(resp.Body, buffer[:readSize])
    fd.Write(buffer[:readSize])

    contentLength -= readSize;
    workerProgress += readSize

    select {
    case progress <- workerProgress:
      workerProgress = 0
    default:
      continue
    }
  }
}

func main() {
  // parse commandline arguments
  flag.Uint64Var(&BUFFER_SIZE, "buffer_size", 8, "buffer size per connections")
  flag.Uint64Var(&N_CONNECTIONS, "connections", 4, "number of connections")
  flag.Parse()
  tail := flag.Args()
  if len(tail) != 1 {
    fmt.Println("Usage: goaxel <url>")
    os.Exit(1)
  }
  url := tail[0]

  // listing information
  fmt.Printf("url: %v\n", url)
  fmt.Printf("buffer_size: %vKB\n", BUFFER_SIZE)
  fmt.Printf("number of connection: %v\n", N_CONNECTIONS)

  // fetch download information
  downloadInfo, err := FetchDownloadInfo(url)
  if err != nil {
    fmt.Println("Error:", err.Error())
    os.Exit(1)
  }

  // create a file a write to that file
  finalFilename, file, err := CreateFile(downloadInfo.OutputFilename)
  if err != nil {
    fmt.Println("Error:", err.Error())
    os.Exit(1)
  }
  defer file.Close()
  // create a file with size as the contentLength
  _, err = file.Seek(int64(downloadInfo.ContentLength - 1), 0)
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
    totalProgress := uint64(0)
    lastSpeed := float64(0)
    
    printProgress := func(speed float64, progress float64) {
      fmt.Printf("\rspeed: %8.2fKBps progress: %6.2f%%", speed, progress)
    }

    for {
      workerProgress, ok := <-progress
      if !ok {
        break
      }

      currentTime := time.Now()
      diffTime := currentTime.Sub(startTime)
      diffMilli := diffTime.Milliseconds()

      totalProgress += workerProgress
      speed := float64(totalProgress) * 1000.0 / float64(diffMilli) / 1024.0
      progressPercent := float64(totalProgress) * 100.0 / float64(downloadInfo.ContentLength)
     
      printProgress(speed, progressPercent)

      lastSpeed = speed
    }
    printProgress(lastSpeed, 100.0)
  }()

  numberOfWorker := uint64(N_CONNECTIONS)
  for i := uint64(0); i < numberOfWorker; i++ {
    start := i * (downloadInfo.ContentLength / numberOfWorker)
    stop := start + (downloadInfo.ContentLength / numberOfWorker) - 1;
    if i == numberOfWorker - 1 {
      stop = downloadInfo.ContentLength - 1;
    }
    workerWg.Add(1)
    go func(start, stop uint64) {
      defer workerWg.Done()
      DownloadRange(progress, downloadInfo, finalFilename, start, stop)
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
