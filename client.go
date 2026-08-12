package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is a minimal HTTP client for the Redash REST API.
type Client struct {
	cfg  Config
	http *http.Client
}

func newClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

func (c *Client) request(method, path string, query url.Values, body any) ([]byte, error) {
	u := c.cfg.URL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Key "+c.cfg.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("redash returned %s: %s", resp.Status, truncate(data, 500))
	}
	return data, nil
}

func (c *Client) Get(path string, query url.Values) (json.RawMessage, error) {
	return c.request(http.MethodGet, path, query, nil)
}

func (c *Client) Post(path string, body any) (json.RawMessage, error) {
	return c.request(http.MethodPost, path, nil, body)
}

func (c *Client) Delete(path string) error {
	_, err := c.request(http.MethodDelete, path, nil, nil)
	return err
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
