package main

import (
	"log"
	"net/http"

	"com.xsir/proxy/internal/config"
	"com.xsir/proxy/internal/httpserver"
	"com.xsir/proxy/internal/httpserver/upstream"
	"github.com/docker/go-units"
)

func main() {

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	afh, err := httpserver.InitAllowedHosts(cfg.AllowedForwardedHosts)
	if err != nil {
		panic("failed to init allowed hosts: " + err.Error())
	}
	ac, err := httpserver.LoadIPAccessControlList(cfg.AllowedSourceIPs)
	if err != nil {
		panic("failed to load alllowed source ips")
	}
	upstreamOptions := upstream.UpstreamOptions{
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
	}
	upstreamClient := upstream.NewUpstreamClient(&upstreamOptions)
	maxBodySize, err := units.FromHumanSize(cfg.MaxBodySize)
	if err != nil {
		panic("failed to parse max body size: " + err.Error())
	}
	serverOptions := &httpserver.ServerOptions{
		MaxBodySize:           maxBodySize,
		UpstreamTimeoutSecond: cfg.UpstreamTimeoutSecond,
		MaxConcurrency:        cfg.MaxProxyConcurrency,
	}
	server := httpserver.NewServer(upstreamClient, serverOptions)
	server.RegisterRoutes(ac, afh)

	log.Println("HTTP Server listen on :" + cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatal(err)
	}

}
