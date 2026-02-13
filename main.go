package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/serve/config"
	"github.com/a-h/serve/handlers"
)

func main() {
	conf, err := config.New()
	if err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}

	if conf.Help {
		conf.FlagSet.PrintDefaults()
		return
	}

	if err = conf.Validate(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	handler, closer, err := handlers.Create(conf)
	if err != nil {
		fmt.Printf("Error creating handler: %v\n", err)
		os.Exit(1)
	}
	defer closer()

	server := &http.Server{
		Addr:              conf.Addr,
		Handler:           handler,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		MaxHeaderBytes:    1 << 20,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	listen := server.ListenAndServe

	serveTLS := conf.Crt != "" && conf.Key != ""
	if serveTLS {
		// Check that we're not attempting to serve the key and crt files.
		absCrt, err := filepath.Abs(conf.Crt)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", conf.Crt, err)
			os.Exit(1)
		}
		absKey, err := filepath.Abs(conf.Key)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", conf.Key, err)
			os.Exit(1)
		}
		absDir, err := filepath.Abs(conf.Dir)
		if err != nil {
			fmt.Printf("Failed to get absolute path of %q: %v\n", conf.Dir, err)
			os.Exit(1)
		}
		if strings.HasPrefix(absCrt, absDir) || strings.HasPrefix(absKey, absDir) {
			fmt.Println("Error: -crt and -key must not be in the directory being served.")
			os.Exit(1)
		}
		// Switch to TLS mode.
		listen = func() error {
			return server.ListenAndServeTLS(conf.Crt, conf.Key)
		}
	}

	fmt.Print(conf.String())
	if err := listen(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
