package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/services"
)

func RateLimitMiddleware(rateLimiter *services.RateLimiter, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			defer cancel()

			var allowed bool
			var remaining int
			var resetTime time.Time
			var err error

			if userID, ok := GetUserID(r); ok {
				// Rate limit by user ID
				allowed, remaining, resetTime, err = rateLimiter.CheckRateLimit(ctx, userID, maxRequests, window)
			} else {
				// Rate limit by IP address
				ip := getClientIP(r)
				allowed, remaining, resetTime, err = rateLimiter.CheckRateLimitByIP(ctx, ip, maxRequests, window)
			}

			if err != nil {
				log.Printf("Rate limiter error: %v", err)
				// On error, allow request (fail open, not fail closed)
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
				http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip := r.RemoteAddr

	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
