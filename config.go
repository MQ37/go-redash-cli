package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds connection settings read from the environment.
type Config struct {
	URL     string
	APIKey  string
	Timeout time.Duration
}

func loadConfig() (Config, error) {
	url := strings.TrimRight(os.Getenv("REDASH_URL"), "/")
	key := os.Getenv("REDASH_API_KEY")
	if url == "" || key == "" {
		return Config{}, fmt.Errorf("REDASH_URL and REDASH_API_KEY must be set")
	}

	timeoutMs := 30000
	if v := os.Getenv("REDASH_TIMEOUT"); v != "" {
		ms, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REDASH_TIMEOUT: %w", err)
		}
		timeoutMs = ms
	}

	return Config{URL: url, APIKey: key, Timeout: time.Duration(timeoutMs) * time.Millisecond}, nil
}
