package httpserver

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"com.xsir/proxy/internal/httpserver/upstream"
)

type Server struct {
	UpstreamClient *upstream.UpstreamClient
	ServerOptions  *ServerOptions
	proxySem       chan struct{}
}

type ServerOptions struct {
	MaxBodySize           int64
	UpstreamTimeoutSecond int
	MaxConcurrency        int
}

func (s *Server) withProxyLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case s.proxySem <- struct{}{}:
			defer func() {
				<-s.proxySem
			}()
			next(w, r)
		default:
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
	}
}

func NewServer(upstreamClient *upstream.UpstreamClient, serverOptions *ServerOptions) *Server {
	return &Server{
		ServerOptions:  serverOptions,
		UpstreamClient: upstreamClient,
		proxySem:       make(chan struct{}, serverOptions.MaxConcurrency),
	}
}

func (c *Server) proxyHander(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	client := r.RemoteAddr
	forwardURL := strings.TrimSpace(r.Header.Get("forward-url"))

	r.Body = http.MaxBytesReader(w, r.Body, c.ServerOptions.MaxBodySize)
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			log.Printf("[proxy] error=body-too-large method=%s client=%s target=%s maxBytes=%d duration=%s",
				r.Method, client, forwardURL, c.ServerOptions.MaxBodySize, time.Since(start))
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		log.Printf("[proxy] error=read-body-failed method=%s client=%s target=%s detail=%v duration=%s",
			r.Method, client, forwardURL, err, time.Since(start))
		http.Error(w, "read body failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	headers := r.Header.Clone()
	headers.Del("forward-url")

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(c.ServerOptions.UpstreamTimeoutSecond))
	defer cancel()

	resp, err := c.UpstreamClient.ForwardRequest(ctx, bodyBytes, forwardURL, headers, r.Method)
	if err != nil {
		log.Printf("[proxy] error=upstream-failed method=%s client=%s target=%s detail=%v duration=%s",
			r.Method, client, forwardURL, err, time.Since(start))
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if err := CopyUpstreamResponse(w, resp); err != nil {
		log.Printf("[proxy] error=copy-response-failed method=%s client=%s target=%s status=%d detail=%v duration=%s",
			r.Method, client, forwardURL, resp.StatusCode, err, time.Since(start))
		log.Println("failed to copy upstream response: " + err.Error())
		return
	}
	log.Printf("[proxy] status=%d method=%s client=%s target=%s duration=%s",
		resp.StatusCode, r.Method, client, forwardURL, time.Since(start))
}

func (c *Server) RegisterRoutes(ipAccessControl *IPAccessControl, allowedHosts map[string]struct{}) {
	var handler http.Handler
	handler = http.HandlerFunc(c.proxyHander)
	handler = Chain(
		handler,
		IPAccessMiddleware(ipAccessControl),
		HostAllowMiddleware(allowedHosts))
	http.HandleFunc("/", c.withProxyLimit(handler.ServeHTTP))
}

func CopyUpstreamResponse(w http.ResponseWriter, resp *http.Response) error {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	_, err := io.Copy(w, resp.Body)
	return err
}
