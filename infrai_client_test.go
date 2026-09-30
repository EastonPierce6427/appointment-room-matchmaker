package appointmentroom

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientDecodesBusinessErrorBeforeStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"INVALID_INPUT","message":"channel is required"},"metadata":null}`))
	}))
	defer server.Close()

	client := NewInfraiClient("test-key")
	client.BaseURL = server.URL
	err := client.post(context.Background(), "/v1/realtime/channel/create", map[string]string{"channel": ""}, "appt-1:channel")
	var apiErr *InfraiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected InfraiError, got %v", err)
	}
	if apiErr.Code != "INVALID_INPUT" || apiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("got code=%q status=%d", apiErr.Code, apiErr.HTTPStatus)
	}
}

func TestClientRetriesRateLimitWithSameIdempotencyKey(t *testing.T) {
	calls := 0
	keys := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		keys[r.Header.Get("Idempotency-Key")] = true
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"RATE_LIMITED","message":"retry later"},"metadata":null}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()

	client := NewInfraiClient("test-key")
	client.BaseURL = server.URL
	client.Sleep = func(context.Context, time.Duration) error { return nil }
	err := client.post(context.Background(), "/v1/realtime/publish", map[string]string{"channel": "appointment-1"}, "appt-1:notification")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(keys) != 1 || !keys["appt-1:notification"] {
		t.Fatalf("calls=%d keys=%v", calls, keys)
	}
}
