package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

func newRateLimiter(requests int, windowMinutes int) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requests,
		window:   time.Duration(windowMinutes) * time.Minute,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if reqs, exists := rl.requests[key]; exists {
		validReqs := []time.Time{}
		for _, reqTime := range reqs {
			if reqTime.After(windowStart) {
				validReqs = append(validReqs, reqTime)
			}
		}
		rl.requests[key] = validReqs
	}

	// Check limit
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add new request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(requests int, windowMinutes int) gin.HandlerFunc {
	limiter := newRateLimiter(requests, windowMinutes)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
