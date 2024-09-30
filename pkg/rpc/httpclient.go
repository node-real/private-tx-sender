package rpc

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   time.Second,
		KeepAlive: 60 * time.Second,
	}

	transport = &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: 2000,
		MaxConnsPerHost:     2000,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	HTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
)

type HTTPCode int

func (code HTTPCode) Success() bool {
	return 200 <= code && code <= 299
}

type CallOptions struct {
	Header map[string]string
}

func (c *CallOptions) ApplyOptions(options ...CallOption) {
	for _, opt := range options {
		opt(c)
	}
}

type CallOption func(*CallOptions)

func WithHeader(header map[string]string) CallOption {
	return func(c *CallOptions) {
		for k, v := range header {
			c.Header[k] = v
		}
	}
}
