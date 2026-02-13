package main

import (
	"crypto/tls"
	"log/slog"
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
		slog.Error("Error parsing config", slog.Any("error", err))
		os.Exit(1)
	}

	if conf.Help {
		conf.FlagSet.PrintDefaults()
		return
	}

	log := createLogger(conf.LogFormat)

	if err = conf.Validate(); err != nil {
		log.Error("Invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	handler, closer, err := handlers.Create(log, conf)
	if err != nil {
		log.Error("Error creating handler", slog.Any("error", err))
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
			log.Error("Failed to get absolute path of crt file", slog.String("crt", conf.Crt), slog.Any("error", err))
			os.Exit(1)
		}
		absKey, err := filepath.Abs(conf.Key)
		if err != nil {
			log.Error("Failed to get absolute path of key file", slog.String("key", conf.Key), slog.Any("error", err))
			os.Exit(1)
		}
		absDir, err := filepath.Abs(conf.Dir)
		if err != nil {
			log.Error("Failed to get absolute path of serve directory", slog.String("dir", conf.Dir), slog.Any("error", err))
			os.Exit(1)
		}
		if strings.HasPrefix(absCrt, absDir) || strings.HasPrefix(absKey, absDir) {
			log.Error("Certificate and key files must not be in the directory being served", slog.String("crt", conf.Crt), slog.String("key", conf.Key), slog.String("dir", conf.Dir))
			os.Exit(1)
		}
		// Switch to TLS mode.
		listen = func() error {
			return server.ListenAndServeTLS(conf.Crt, conf.Key)
		}
	}

	log.Info("Starting server", slog.String("dir", conf.Dir), slog.String("addr", conf.Addr), slog.Bool("tls", serveTLS), slog.Bool("log-remote-addr", conf.LogRemoteAddr), slog.Bool("read-only", conf.ReadOnly), slog.Bool("auth-enabled", conf.Auth != ""))

	if err := listen(); err != nil {
		log.Error("Server error", slog.Any("error", err))
		os.Exit(1)
	}
}

func createLogger(logFormat string) *slog.Logger {
	if logFormat == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
