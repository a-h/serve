package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var flagDir = flag.String("dir", ".", "Directory to serve.")
var flagAddr = flag.String("addr", ":8080", "Address to serve on.")
var flagCrt = flag.String("crt", "", "Path to crt file.")
var flagKey = flag.String("key", "", "Path to key file.")
var flagHelp = flag.Bool("help", false, "Print help.")

func main() {
	flag.Parse()
	if *flagHelp {
		flag.PrintDefaults()
		return
	}
	if *flagCrt != "" && *flagKey == "" || *flagCrt == "" && *flagKey != "" {
		fmt.Println("Error: -crt and -key must be used together.")
		os.Exit(1)
	}

	fs := http.FileServer(http.Dir(*flagDir))
	server := &http.Server{
		Addr: *flagAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s %v %v\n", time.Now().Format(time.RFC3339), r.Method, r.URL.String())
			fs.ServeHTTP(w, r)
		}),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	listen := server.ListenAndServe

	serveTLS := *flagCrt != "" && *flagKey != ""
	if serveTLS {
		// Check that we're not attempting to serve the key and crt files.
		absCrt, err := filepath.Abs(*flagCrt)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", *flagCrt, err)
			os.Exit(1)
		}
		absKey, err := filepath.Abs(*flagKey)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", *flagKey, err)
			os.Exit(1)
		}
		absDir, err := filepath.Abs(*flagDir)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", *flagDir, err)
			os.Exit(1)
		}
		if strings.HasPrefix(absCrt, absDir) || strings.HasPrefix(absKey, absDir) {
			fmt.Println("Error: -crt and -key must not be in the directory being served.")
			os.Exit(1)
		}
		// Switch to TLS mode.
		listen = func() error {
			return server.ListenAndServeTLS(*flagCrt, *flagKey)
		}
	}

	fmt.Printf("Serving %q on %s\n", *flagDir, *flagAddr)
	if err := listen(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
