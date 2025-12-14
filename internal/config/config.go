package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr            string   `yaml:"listen_addr"`
	LogLevel              string   `yaml:"log_level"`
	UpstreamTimeoutSecond int      `yaml:"upstream_time_seconds"`
	AllowedForwardedHosts []string `yaml:"allowed_forwarded_hosts"`
	AllowedSourceIPs      []string `yaml:"allowed_source_ips"`
	MaxIdleConns          int      `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int      `yaml:"max_idle_comms_per_host"`
	IdleConnTimeout       int      `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   int      `yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout int      `yaml:"expect_continue_timeout"`
	ResponseHeaderTimeout int      `yaml:"response_header_timeout"`
	MaxBodySize           string   `yaml:"max_body_size"`
	MaxProxyConcurrency   int      `yaml:"max_proxy_concurrency"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
