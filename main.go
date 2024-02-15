package main

import (
	"flag"
	"fmt"
	"goaxel/goaxel"
	"os"
)

var (
	BUFFER_SIZE   uint64
	N_CONNECTIONS uint64
)

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

	// download file
	goaxel.Download(N_CONNECTIONS, BUFFER_SIZE, url)
}
