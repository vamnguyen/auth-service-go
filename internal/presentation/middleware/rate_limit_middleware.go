package middleware

import (
	"sync"
	"time"

	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count      int
	resetAt    time.Time
}

func NewRateLimiter(requestsPerMinute int) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.After(info.resetAt) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		info, exists := rl.requests[clientIP]

		if !exists || now.After(info.resetAt) {
			rl.requests[clientIP] = &clientInfo{
				count:   1,
				resetAt: now.Add(rl.window),
			}
			c.Next()
			return
		}

		if info.count >= rl.limit {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later")
			c.Abort()
			return
		}

		info.count++
		c.Next()
	}
}
