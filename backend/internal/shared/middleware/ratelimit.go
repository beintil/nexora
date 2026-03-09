package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimitRule describes a rate limit: MaxRequests per Window for paths starting with PathPrefix.
type RateLimitRule struct {
	PathPrefix  string
	MaxRequests int
	Window      time.Duration
}

// RateLimiter returns middleware that enforces rate-limit rules using Redis sliding window counters.
// Each rule matches requests whose path starts with the given prefix.
// Key format: "rl:<ip>:<pathPrefix>:<window_bucket>".
// When the limit is exceeded, the middleware responds with 429 Too Many Requests.
func RateLimiter(redisClient *redis.Client, rules []RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			ip := extractIP(r)

			for _, rule := range rules {
				if !strings.HasPrefix(path, rule.PathPrefix) {
					continue
				}

				bucket := time.Now().Unix() / int64(rule.Window.Seconds())
				key := fmt.Sprintf("rl:%s:%s:%d", ip, rule.PathPrefix, bucket)

				ctx := r.Context()
				count, err := redisClient.Incr(ctx, key).Result()
				if err != nil {
					// On Redis error, allow the request to proceed (fail-open)
					break
				}
				if count == 1 {
					redisClient.Expire(ctx, key, rule.Window+time.Second)
				}

				if count > int64(rule.MaxRequests) {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rule.Window.Seconds())))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, `{"error":"rate_limit_exceeded","message":"Too many requests. Try again later.","code":429}`)
					return
				}

				// First matching rule wins
				break
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitHandler returns middleware that enforces a rate-limit rule using a sliding window counter.
// Key format: "rl:<ip>:<path>:<window_bucket>".
func RateLimitHandler(redisClient *redis.Client, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			ip := extractIP(r)
			bucket := time.Now().Unix() / int64(window.Seconds())
			key := fmt.Sprintf("rl:%s:%s:%d", ip, path, bucket)

			ctx := r.Context()
			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				// On Redis error, allow the request to proceed (fail-open)
				next.ServeHTTP(w, r)
				return
			}
			if count == 1 {
				redisClient.Expire(ctx, key, window+time.Second)
			}

			if count > int64(maxRequests) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"rate_limit_exceeded","message":"Too many requests. Try again later.","code":429}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractIP returns the client IP from X-Forwarded-For, X-Real-IP, or RemoteAddr.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
