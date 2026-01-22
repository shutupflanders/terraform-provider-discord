package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	maxRetries  = 5
	baseBackoff = 1 * time.Second
	maxBackoff  = 120 * time.Second
)

// RateLimitError represents a Discord rate limit response
type RateLimitError struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

// executeWithRetry wraps discordgo API calls with rate limit handling
func executeWithRetry[T any](ctx context.Context, operation func() (T, error)) (T, error) {
	var result T
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		result, err = operation()

		if err == nil {
			return result, nil
		}

		// Check if it's a rate limit error
		if restErr, ok := err.(*discordgo.RESTError); ok {
			if restErr.Response != nil && restErr.Response.StatusCode == 429 {
				retryAfter := parseRetryAfter(restErr)
				waitDuration := calculateBackoff(attempt, retryAfter)

				// Wait with context cancellation support
				select {
				case <-ctx.Done():
					return result, ctx.Err()
				case <-time.After(waitDuration):
					continue
				}
			}
		}

		// Not a rate limit error, return immediately
		return result, err
	}

	return result, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
}

// executeWithRetryNoResult wraps discordgo API calls that return only an error
func executeWithRetryNoResult(ctx context.Context, operation func() error) error {
	_, err := executeWithRetry(ctx, func() (struct{}, error) {
		return struct{}{}, operation()
	})
	return err
}

// parseRetryAfter extracts the retry duration from a 429 response
func parseRetryAfter(restErr *discordgo.RESTError) time.Duration {
	if restErr.Response == nil {
		return 5 * time.Second
	}

	// Try Retry-After header first
	if retryAfterStr := restErr.Response.Header.Get("Retry-After"); retryAfterStr != "" {
		if seconds, err := strconv.ParseFloat(retryAfterStr, 64); err == nil {
			return time.Duration(seconds * float64(time.Second))
		}
	}

	// Try parsing JSON body from the response body in the error
	if restErr.ResponseBody != nil {
		var rateLimitErr RateLimitError
		if json.Unmarshal(restErr.ResponseBody, &rateLimitErr) == nil && rateLimitErr.RetryAfter > 0 {
			return time.Duration(rateLimitErr.RetryAfter * float64(time.Second))
		}
	}

	// Try reading from Response.Body if ResponseBody is nil
	if restErr.Response.Body != nil {
		body, err := io.ReadAll(restErr.Response.Body)
		if err == nil {
			var rateLimitErr RateLimitError
			if json.Unmarshal(body, &rateLimitErr) == nil && rateLimitErr.RetryAfter > 0 {
				return time.Duration(rateLimitErr.RetryAfter * float64(time.Second))
			}
		}
	}

	// Default fallback
	return 5 * time.Second
}

// calculateBackoff returns wait duration with exponential backoff
func calculateBackoff(attempt int, retryAfter time.Duration) time.Duration {
	// Use Retry-After if provided and reasonable
	if retryAfter > 0 && retryAfter < maxBackoff {
		// Add small buffer
		return retryAfter + 500*time.Millisecond
	}

	// Exponential backoff: 1s, 2s, 4s, 8s, 16s...
	backoff := baseBackoff * time.Duration(1<<uint(attempt))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}
