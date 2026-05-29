package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rate:    r,
		burst:   burst,
	}

	// Cleanup stale clients every minute
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		c = &client{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.clients[ip] = c
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func (rl *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			utils.HandleError(c, &apperrors.TooManyRequestsError{Message: "too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
