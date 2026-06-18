package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements a sliding window rate limiter using Valkey/Redis
type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter creates a new RateLimiter instance
func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Handler returns a middleware handler for rate limiting
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.client == nil {
			next.ServeHTTP(w, r)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		ctx := r.Context()
		now := time.Now().UnixNano()
		clearBefore := time.Now().Add(-rl.window).UnixNano()
		key := fmt.Sprintf("rate_limit:%s", ip)

		// Pipeline execution to clean old requests, add new request, and count active requests within the window
		pipe := rl.client.TxPipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(clearBefore, 10))
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: strconv.FormatInt(now, 10)})
		pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, rl.window)

		cmds, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Warning: rate limiter Valkey command failed: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		// ZCard command is the 3rd operation (index 2)
		cardCmd, ok := cmds[2].(*redis.IntCmd)
		if !ok {
			log.Printf("Warning: rate limiter failed to parse card result")
			next.ServeHTTP(w, r)
			return
		}

		count, err := cardCmd.Result()
		if err != nil {
			log.Printf("Warning: rate limiter failed to get card result: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		if int(count) > rl.limit {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.limit-int(count)))

		next.ServeHTTP(w, r)
	})
}
