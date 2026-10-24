package provider

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryableClientRetriesRateLimiting(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limited for the first 6 requests, more than the default 4 retries
		if calls.Add(1) <= 6 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newRetryableClient(retryMax, time.Millisecond, 5*time.Millisecond)
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK || calls.Load() != 7 {
		t.Errorf("expected success after 7 calls, got status %d after %d calls", resp.StatusCode, calls.Load())
	}
}

func TestBackoffWithJitter(t *testing.T) {
	for attempt := 0; attempt < 12; attempt++ {
		sleep := backoffWithJitter(retryWaitMin, retryWaitMax, attempt, nil)
		if sleep < retryWaitMin || sleep > retryWaitMax {
			t.Errorf("attempt %d: backoff %s outside [%s, %s]", attempt, sleep, retryWaitMin, retryWaitMax)
		}
	}

	resp := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"7"}}}
	if sleep := backoffWithJitter(retryWaitMin, retryWaitMax, 0, resp); sleep != 7*time.Second {
		t.Errorf("expected Retry-After to be honoured, got %s", sleep)
	}
}
