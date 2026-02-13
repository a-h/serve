package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

func New() (c *Config, err error) {
	conf := &Config{
		Dir:           ".",
		Addr:          ":8080",
		Crt:           "",
		Key:           "",
		LogRemoteAddr: false,
		ReadOnly:      true,
		Auth:          "",
		LogFormat:     "text",
		Help:          false,
	}

	conf.FlagSet = flag.NewFlagSet("serve", flag.ContinueOnError)
	conf.FlagSet.StringVar(&conf.Dir, "dir", conf.Dir, "Directory to serve. (Env: SERVE_DIR)")
	conf.FlagSet.StringVar(&conf.Addr, "addr", conf.Addr, "Address to serve on. (Env: SERVE_ADDR)")
	conf.FlagSet.StringVar(&conf.Crt, "crt", conf.Crt, "Path to crt file for TLS. (Env: SERVE_CRT)")
	conf.FlagSet.StringVar(&conf.Key, "key", conf.Key, "Path to key file for TLS. (Env: SERVE_KEY)")
	conf.FlagSet.BoolVar(&conf.LogRemoteAddr, "log-remote-addr", conf.LogRemoteAddr, "Log remote address. (Env: SERVE_LOG_REMOTE_ADDR)")
	conf.FlagSet.BoolVar(&conf.ReadOnly, "read-only", conf.ReadOnly, "Allow only GET and HEAD requests. (Env: SERVE_READ_ONLY)")
	conf.FlagSet.StringVar(&conf.Auth, "auth", conf.Auth, "Username:Password for basic auth, no auth if not set. (Env: SERVE_AUTH)")
	conf.FlagSet.DurationVar(&conf.ReadTimeout, "read-timeout", 24*time.Hour, "Maximum duration for reading the entire request, including the body. (Env: SERVE_READ_TIMEOUT)")
	conf.FlagSet.DurationVar(&conf.ReadHeaderTimeout, "read-header-timeout", 5*time.Second, "Amount of time allowed to read request headers. (Env: SERVE_READ_HEADER_TIMEOUT)")
	conf.FlagSet.DurationVar(&conf.WriteTimeout, "write-timeout", 12*time.Hour, "Maximum duration before timing out writes of the response. (Env: SERVE_WRITE_TIMEOUT)")
	conf.FlagSet.StringVar(&conf.LogFormat, "log-format", conf.LogFormat, "Log format: text or json. (Env: SERVE_LOG_FORMAT)")
	conf.FlagSet.BoolVar(&conf.Help, "help", conf.Help, "Print help.")
	if err = conf.FlagSet.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// Parse environment variables, which take precedence over flags.
	var errs []error
	if dirEnv := os.Getenv("SERVE_DIR"); dirEnv != "" {
		conf.Dir = dirEnv
	}
	if addrEnv := os.Getenv("SERVE_ADDR"); addrEnv != "" {
		conf.Addr = addrEnv
	}
	if crtEnv := os.Getenv("SERVE_CRT"); crtEnv != "" {
		conf.Crt = crtEnv
	}
	if keyEnv := os.Getenv("SERVE_KEY"); keyEnv != "" {
		conf.Key = keyEnv
	}
	if remoteAddrEnv := os.Getenv("SERVE_LOG_REMOTE_ADDR"); remoteAddrEnv != "" {
		conf.LogRemoteAddr = remoteAddrEnv == "true"
	}
	if readOnlyEnv := os.Getenv("SERVE_READ_ONLY"); readOnlyEnv != "" {
		conf.ReadOnly = readOnlyEnv == "true"
	}
	if authEnv := os.Getenv("SERVE_AUTH"); authEnv != "" {
		conf.Auth = authEnv
	}
	conf.ReadTimeout, err = parseDurationEnv("SERVE_READ_TIMEOUT", conf.ReadTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid SERVE_READ_TIMEOUT: %w", err))
	}
	conf.ReadHeaderTimeout, err = parseDurationEnv("SERVE_READ_HEADER_TIMEOUT", conf.ReadHeaderTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid SERVE_READ_HEADER_TIMEOUT: %w", err))
	}
	conf.WriteTimeout, err = parseDurationEnv("SERVE_WRITE_TIMEOUT", conf.WriteTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid SERVE_WRITE_TIMEOUT: %w", err))
	}
	conf.LogFormat, err = parseLogFormat("SERVE_LOG_FORMAT", conf.LogFormat)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid SERVE_LOG_FORMAT: %w", err))
	}

	return conf, errors.Join(errs...)
}

func parseLogFormat(envVar string, defaultVal string) (string, error) {
	val := os.Getenv(envVar)
	if val == "" {
		return defaultVal, nil
	}
	if val != "text" && val != "json" {
		return "", fmt.Errorf("invalid log format %q, allowed values are: text, json", val)
	}
	return val, nil
}

func parseDurationEnv(envVar string, defaultVal time.Duration) (d time.Duration, err error) {
	val := os.Getenv(envVar)
	if val == "" {
		return defaultVal, nil
	}
	return time.ParseDuration(val)
}

type Config struct {
	FlagSet           *flag.FlagSet
	Dir               string
	Addr              string
	Crt               string
	Key               string
	LogRemoteAddr     bool
	ReadOnly          bool
	Auth              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	LogFormat         string
	Help              bool
}

func (c *Config) Validate() error {
	if (c.Crt != "" && c.Key == "") || (c.Crt == "" && c.Key != "") {
		return ErrCrtKeyMismatch
	}
	return nil
}

var ErrCrtKeyMismatch = fmt.Errorf("-crt and -key must be used together.")
