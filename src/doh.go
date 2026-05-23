package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DoHClient struct {
	cfg     *Config
	client  *http.Client
	fullURL string
}

func NewDoHClient(cfg *Config) (*DoHClient, error) {
	base, err := url.Parse(cfg.Upstream.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream.url: %w", err)
	}
	if base.Scheme != "https" && base.Scheme != "http" {
		return nil, fmt.Errorf("upstream.url scheme must be http or https")
	}

	full := strings.TrimRight(cfg.Upstream.URL, "/")
	if p := strings.TrimSpace(cfg.Upstream.Path); p != "" {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		full += p
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.Upstream.InsecureSkipVerify,
		ServerName:         cfg.Upstream.ServerName,
	}
	transport := &http.Transport{
		ForceAttemptHTTP2:   true,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConns:        64,
		MaxIdleConnsPerHost: 32,
		TLSClientConfig:     tlsCfg,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Upstream.Timeout.Std(),
	}
	return &DoHClient{cfg: cfg, client: client, fullURL: full}, nil
}

func (c *DoHClient) Query(ctx context.Context, msg []byte) ([]byte, error) {
	var (
		req *http.Request
		err error
	)
	if c.cfg.Upstream.Method == "GET" {
		encoded := base64.RawURLEncoding.EncodeToString(msg)
		sep := "?"
		if strings.Contains(c.fullURL, "?") {
			sep = "&"
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, c.fullURL+sep+"dns="+encoded, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL, bytes.NewReader(msg))
		if err == nil {
			req.Header.Set("Content-Type", "application/dns-message")
		}
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-message")
	for k, v := range c.cfg.Upstream.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
