package proxy

import (
	"errors"
	"net/url"
)

var (
	ErrInvalidScheme = errors.New("invalid scheme")
	ErrInvalidHost   = errors.New("invalid host")
)

func IsValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func ParseHTTPURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrInvalidScheme
	}

	if u.Host == "" {
		return nil, ErrInvalidHost
	}

	return u, nil
}
