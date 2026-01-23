package discord

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestParseRetryAfter_Header(t *testing.T) {
	restErr := &discordgo.RESTError{
		Response: &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": []string{"5"}},
		},
	}

	duration := parseRetryAfter(restErr)
	expected := 5 * time.Second
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestParseRetryAfter_HeaderFloat(t *testing.T) {
	restErr := &discordgo.RESTError{
		Response: &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": []string{"1.5"}},
		},
	}

	duration := parseRetryAfter(restErr)
	expected := 1500 * time.Millisecond
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestParseRetryAfter_JSONBody(t *testing.T) {
	restErr := &discordgo.RESTError{
		Response: &http.Response{
			StatusCode: 429,
			Header:     http.Header{},
		},
		ResponseBody: []byte(`{"message": "You are being rate limited.", "retry_after": 3.0, "global": false}`),
	}

	duration := parseRetryAfter(restErr)
	expected := 3 * time.Second
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestParseRetryAfter_Fallback(t *testing.T) {
	restErr := &discordgo.RESTError{
		Response: &http.Response{
			StatusCode: 429,
			Header:     http.Header{},
		},
	}

	duration := parseRetryAfter(restErr)
	expected := 5 * time.Second
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestParseRetryAfter_NilResponse(t *testing.T) {
	restErr := &discordgo.RESTError{
		Response: nil,
	}

	duration := parseRetryAfter(restErr)
	expected := 5 * time.Second
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestCalculateBackoff_WithRetryAfter(t *testing.T) {
	duration := calculateBackoff(0, 3*time.Second)
	expected := 3*time.Second + 500*time.Millisecond
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestCalculateBackoff_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			duration := calculateBackoff(tt.attempt, 0)
			if duration != tt.expected {
				t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, duration)
			}
		})
	}
}

func TestCalculateBackoff_CapsAtMax(t *testing.T) {
	duration := calculateBackoff(10, 0)
	if duration != maxBackoff {
		t.Errorf("expected max backoff %v, got %v", maxBackoff, duration)
	}
}

func TestCalculateBackoff_IgnoresExcessiveRetryAfter(t *testing.T) {
	// If Retry-After exceeds maxBackoff, fall through to exponential backoff
	duration := calculateBackoff(0, 200*time.Second)
	expected := 1 * time.Second // attempt 0 exponential backoff
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	calls := 0

	result, err := executeWithRetry(ctx, func() (string, error) {
		calls++
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("expected 'ok', got '%s'", result)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestExecuteWithRetry_NonRateLimitError(t *testing.T) {
	ctx := context.Background()
	calls := 0

	_, err := executeWithRetry(ctx, func() (string, error) {
		calls++
		return "", fmt.Errorf("some other error")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry for non-rate-limit errors), got %d", calls)
	}
}

func TestExecuteWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executeWithRetry(ctx, func() (string, error) {
		return "ok", nil
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestExecuteWithRetryNoResult_Success(t *testing.T) {
	ctx := context.Background()
	calls := 0

	err := executeWithRetryNoResult(ctx, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestExecuteWithRetryNoResult_Error(t *testing.T) {
	ctx := context.Background()

	err := executeWithRetryNoResult(ctx, func() error {
		return fmt.Errorf("failed")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
