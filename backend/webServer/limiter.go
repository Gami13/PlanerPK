package webServer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimitPolicy struct {
	Name        string
	Requests    int
	Window      time.Duration
	BanDuration time.Duration
}

type RateLimitDecision struct {
	Allowed     bool
	RetryAfter  time.Duration
	BannedUntil time.Time
}

type clientTracker struct {
	mu          sync.Mutex
	events      []time.Time
	bannedUntil time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	trackers map[string]*clientTracker
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{trackers: make(map[string]*clientTracker)}
}

func (rl *RateLimiter) Evaluate(key string, policy RateLimitPolicy, now time.Time) RateLimitDecision {
	if key == "" || policy.Requests <= 0 {
		return RateLimitDecision{Allowed: true}
	}

	tracker := rl.getTracker(key)

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	if !tracker.bannedUntil.IsZero() && now.Before(tracker.bannedUntil) {
		return RateLimitDecision{
			Allowed:     false,
			RetryAfter:  tracker.bannedUntil.Sub(now),
			BannedUntil: tracker.bannedUntil,
		}
	}

	cutoff := now.Add(-policy.Window)
	valid := tracker.events[:0]
	for _, ts := range tracker.events {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	tracker.events = valid

	if len(tracker.events) >= policy.Requests {
		decision := RateLimitDecision{Allowed: false}
		if policy.BanDuration > 0 {
			tracker.bannedUntil = now.Add(policy.BanDuration)
			decision.BannedUntil = tracker.bannedUntil
			decision.RetryAfter = policy.BanDuration
			return decision
		}

		oldest := tracker.events[0]
		retryIn := oldest.Add(policy.Window).Sub(now)
		if retryIn < 0 {
			retryIn = 0
		}
		decision.RetryAfter = retryIn
		return decision
	}

	tracker.events = append(tracker.events, now)
	return RateLimitDecision{Allowed: true}
}

func (rl *RateLimiter) getTracker(key string) *clientTracker {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tracker, exists := rl.trackers[key]
	if !exists {
		tracker = &clientTracker{}
		rl.trackers[key] = tracker
	}
	return tracker
}

func rateLimitMiddleware(limiter *RateLimiter, policy RateLimitPolicy, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := composeRateLimitKey(r, policy.Name)
		decision := limiter.Evaluate(key, policy, time.Now())
		if decision.Allowed {
			next.ServeHTTP(w, r)
			return
		}

		message := buildRateLimitMessage(decision)
		slog.Warn("Request blocked by rate limiter",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("key", key),
			slog.String("policy", policy.Name),
			slog.Duration("window", policy.Window),
			slog.Int("limit", policy.Requests),
			slog.Duration("retry_after", decision.RetryAfter))

		writeRateLimitResponse(w, decision, message)
	})
}

func composeRateLimitKey(r *http.Request, policyName string) string {
	token := r.Header.Get("Authorization")
	if token != "" {
		return fmt.Sprintf("%s|policy:%s", baseRateLimitKey(r, "token:"+token), policyName)
	}

	// No IP-based rate limiting: behind Cloudflare all requests share the
	// same remote address, so per-IP limits would throttle every user.
	return ""
}

func baseRateLimitKey(r *http.Request, identifier string) string {
	return fmt.Sprintf("%s %s|%s", r.Method, r.URL.Path, identifier)
}

func buildRateLimitMessage(decision RateLimitDecision) string {
	if !decision.BannedUntil.IsZero() {
		return "Rate limit exceeded for this endpoint. Access blocked for one week or until the server restarts."
	}
	seconds := int(decision.RetryAfter.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", seconds)
}

func writeRateLimitResponse(w http.ResponseWriter, decision RateLimitDecision, message string) {
	w.Header().Set("Content-Type", "application/json")
	if decision.RetryAfter > 0 {
		retrySeconds := int(decision.RetryAfter / time.Second)
		if decision.RetryAfter%time.Second != 0 {
			retrySeconds++
		}
		w.Header().Set("Retry-After", strconv.Itoa(retrySeconds))
	}

	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
