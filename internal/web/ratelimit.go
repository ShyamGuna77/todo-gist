package web

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (app *Application) rateLimit(next http.Handler) http.Handler {
	// If not configured, do nothing.
	if !app.RateLimitEnabled {
		return next
	}
	if app.RateLimitRPS <= 0 || app.RateLimitBurst <= 0 {
		return next
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Cleanup old client entries periodically.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		mu.Lock()
		c, ok := clients[ip]
		if !ok {
			c = &client{
				limiter: rate.NewLimiter(rate.Limit(app.RateLimitRPS), app.RateLimitBurst),
			}
			clients[ip] = c
		}
		c.lastSeen = time.Now()
		allow := c.limiter.Allow()
		mu.Unlock()

		if !allow {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

