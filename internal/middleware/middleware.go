// Package middleware provides cross-cutting HTTP middleware: anonymous
// server-issued sessions, security headers, per-session rate limiting, and
// admin authentication.
package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// SessionCookieName is the name of the signed, httpOnly session cookie.
	SessionCookieName = "w2w_session"
	// userIDKey is the gin context key under which the resolved user ID lives.
	userIDKey = "user_id"
	// sessionMaxAge is how long a session cookie remains valid (1 year).
	sessionMaxAge = 365 * 24 * 60 * 60
)

// GetUserID returns the session-derived user ID for the current request.
// It is only populated after the Session middleware has run.
func GetUserID(c *gin.Context) string {
	if v, ok := c.Get(userIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// Session resolves an opaque, server-issued anonymous identity from a signed
// httpOnly cookie. If no valid cookie is present a fresh identity is minted
// and set. The identity is stored in the gin context and must be read via
// GetUserID — it is never trusted from the request body or query string.
func Session(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		var userID string

		if raw, err := c.Cookie(SessionCookieName); err == nil {
			if id, ok := verifyToken(raw, key); ok {
				userID = id
			}
		}

		if userID == "" {
			userID = newID()
			setSessionCookie(c, signToken(userID, key))
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

// newID returns a 128-bit opaque random identifier as hex.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure is not recoverable in a meaningful way; fall back
		// to a time-seeded value so the request can still proceed.
		return "u" + hex.EncodeToString([]byte(time.Now().String()))[:32]
	}
	return hex.EncodeToString(b)
}

// signToken returns "<id>.<hmac>" where hmac is HMAC-SHA256(id) hex-encoded.
func signToken(id string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id))
	return id + "." + hex.EncodeToString(mac.Sum(nil))
}

// verifyToken validates a signed token and returns the embedded id.
func verifyToken(token string, key []byte) (string, bool) {
	idx := strings.LastIndex(token, ".")
	if idx <= 0 || idx == len(token)-1 {
		return "", false
	}
	id, sig := token[:idx], token[idx+1:]

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id))
	expected := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		return "", false
	}
	return id, true
}

// setSessionCookie writes the signed session cookie. It is marked Secure when
// the request arrived over HTTPS (directly, or via a proxy that set
// X-Forwarded-Proto), so it still works for plain-HTTP local development.
func setSessionCookie(c *gin.Context, value string) {
	secure := c.Request.TLS != nil ||
		strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, value, sessionMaxAge, "/", "", secure, true)
}

// SecurityHeaders applies baseline hardening headers in-app, so the app is not
// dependent on an external proxy being correctly configured.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// The SPA uses inline styles (Tailwind/framer-motion) and talks only to
		// its own origin; keep the policy strict but compatible with that.
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"img-src 'self' data: https:; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src 'self' https://fonts.gstatic.com data:; "+
				"connect-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'")
		c.Next()
	}
}

// AdminAuth gates routes behind a shared secret supplied in the X-Admin-Secret
// header. If no secret is configured the routes are disabled entirely.
func AdminAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Admin endpoints are disabled (ADMIN_SECRET not set)",
			})
			return
		}
		provided := c.GetHeader("X-Admin-Secret")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

// rateLimiter is a fixed-window, in-memory limiter keyed by an arbitrary string
// (session ID, falling back to client IP).
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	counters map[string]*windowCounter
}

type windowCounter struct {
	count       int
	windowStart time.Time
}

// RateLimit returns middleware allowing at most perMinute requests per key in
// any rolling one-minute fixed window. A value <= 0 disables limiting.
func RateLimit(perMinute int) gin.HandlerFunc {
	if perMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	rl := &rateLimiter{
		limit:    perMinute,
		window:   time.Minute,
		counters: make(map[string]*windowCounter),
	}
	rl.startJanitor()

	return func(c *gin.Context) {
		key := GetUserID(c)
		if key == "" {
			key = "ip:" + c.ClientIP()
		}

		if !rl.allow(key) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please slow down.",
			})
			return
		}
		c.Next()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	wc, ok := rl.counters[key]
	if !ok || now.Sub(wc.windowStart) >= rl.window {
		rl.counters[key] = &windowCounter{count: 1, windowStart: now}
		return true
	}

	if wc.count >= rl.limit {
		return false
	}
	wc.count++
	return true
}

// startJanitor periodically evicts stale counters so the map cannot grow
// without bound.
func (rl *rateLimiter) startJanitor() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			rl.mu.Lock()
			for k, wc := range rl.counters {
				if now.Sub(wc.windowStart) >= rl.window {
					delete(rl.counters, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
