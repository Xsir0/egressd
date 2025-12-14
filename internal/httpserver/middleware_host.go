package httpserver

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func InitAllowedHosts(hosts []string) (map[string]struct{}, error) {
	allowedHosts := make(map[string]struct{})
	if len(hosts) == 0 {
		log.Println("[warn] allowed_hosts is empty")
		return allowedHosts, nil
	}

	for _, host := range hosts {
		h := strings.TrimSpace(host)
		if h == "" {
			continue
		}
		allowedHosts[h] = struct{}{}
	}
	return allowedHosts, nil
}

func HostAllowMiddleware(allowedHost map[string]struct{}) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			forwardURL := r.Header.Get("forward-url")
			if forwardURL == "" {
				http.Error(w, "forward is required", http.StatusBadRequest)
				return
			}
			u, err := url.Parse(forwardURL)
			if err != nil {
				http.Error(w, "invalid forward-url", http.StatusBadRequest)
				return
			}

			host := strings.ToLower(u.Hostname())
			if len(allowedHost) > 0 {
				if _, ok := allowedHost[host]; !ok {
					http.Error(w, "host not allowed", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
