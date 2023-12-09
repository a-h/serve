package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var flagDir = flag.String("dir", ".", "Directory to serve.")
var flagAddr = flag.String("addr", ":8080", "Address to serve on.")
var flagHelp = flag.Bool("help", false, "Print help.")

func main() {
	flag.Parse()
	if *flagHelp {
		flag.PrintDefaults()
		return
	}
	http.Handle("/", http.FileServer(http.Dir(*flagDir)))
	fmt.Printf("Serving %q on %s\n", *flagDir, *flagAddr)
	if err := http.ListenAndServe(*flagAddr, nil); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
