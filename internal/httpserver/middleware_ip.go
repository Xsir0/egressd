package httpserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type IPAccessControl struct {
	singleIPs map[string]struct{}
	cidrs     []*net.IPNet
}

func IPAccessMiddleware(ac *IPAccessControl) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ac.IsEmpty() {
				next.ServeHTTP(w, r)
				return
			}
			clientIP, err := GetClientIP(r)
			if err != nil {
				http.Error(w, "invalid client ip", http.StatusBadRequest)
				return
			}
			if !ac.Allow(clientIP) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (ac *IPAccessControl) IsEmpty() bool {
	return len(ac.cidrs) == 0 && len(ac.singleIPs) == 0
}

func (ac *IPAccessControl) Allow(ip net.IP) bool {
	if _, ok := ac.singleIPs[ip.String()]; ok {
		return true
	}
	for _, cidr := range ac.cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func GetClientIP(r *http.Request) (net.IP, error) {
	// 1. X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		ipStr := strings.TrimSpace(parts[0])
		ip := net.ParseIP(ipStr)
		if ip != nil {
			return ip, nil
		}
	}

	// 2. X-Real-IP
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		ip := net.ParseIP(strings.TrimSpace(xrip))
		if ip != nil {
			return ip, nil
		}
	}

	// 3. RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip: %s", host)
	}
	return ip, nil
}

func LoadIPAccessControlList(rules []string) (*IPAccessControl, error) {
	ac := &IPAccessControl{
		singleIPs: make(map[string]struct{}),
		cidrs:     make([]*net.IPNet, 0),
	}

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if strings.Contains(rule, "/") {
			_, ipNet, err := net.ParseCIDR(rule)
			if err != nil {
				return nil, fmt.Errorf("invalid cidr: %s", rule)
			}
			ac.cidrs = append(ac.cidrs, ipNet)
			continue
		}
		ip := net.ParseIP(rule)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip: %s", rule)
		}
		ac.singleIPs[ip.String()] = struct{}{}
	}
	return ac, nil
}
