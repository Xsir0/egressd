package upstream

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"time"
)

type UpstreamClient struct {
	HTTPClient *http.Client
}

type UpstreamOptions struct {
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       int
	TLSHandshakeTimeout   int
	ExpectContinueTimeout int
	ResponseHeaderTimeout int
}

func (c *UpstreamClient) ForwardRequest(ctx context.Context, body []byte, targetUrl string, headers http.Header, method string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetUrl, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = headers.Clone()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func NewUpstreamClient(config *UpstreamOptions) *UpstreamClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       time.Duration(config.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(config.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(config.ExpectContinueTimeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(config.ResponseHeaderTimeout) * time.Second,
	}
	return &UpstreamClient{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   0,
		},
	}
}
