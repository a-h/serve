package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
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
	fs := http.FileServer(http.Dir(*flagDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %v %v\n", time.Now().Format(time.RFC3339), r.Method, r.URL.String())
		fs.ServeHTTP(w, r)
	})
	fmt.Printf("Serving %q on %s\n", *flagDir, *flagAddr)
	if err := http.ListenAndServe(*flagAddr, nil); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
