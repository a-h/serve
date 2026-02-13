package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/serve/handlers"
)

var flagDir = flag.String("dir", ".", "Directory to serve.")
var flagAddr = flag.String("addr", ":8080", "Address to serve on.")
var flagCrt = flag.String("crt", "", "Path to crt file.")
var flagKey = flag.String("key", "", "Path to key file.")
var flagRemoteAddr = flag.Bool("remote-addr", false, "Log remote address.")
var flagHelp = flag.Bool("help", false, "Print help.")
var flagWritable = flag.Bool("writable", false, "Allow POST, PUT, DELETE methods.")
var flagAuth = flag.String("auth", "", "Username:Password for basic auth, no auth if not set.")

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

	handler, closer, err := handlers.Create(*flagDir, *flagRemoteAddr, *flagWritable, *flagAuth)
	if err != nil {
		fmt.Printf("Error creating handler: %v\n", err)
		os.Exit(1)
	}
	defer closer()

	server := &http.Server{
		Addr:           *flagAddr,
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
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
