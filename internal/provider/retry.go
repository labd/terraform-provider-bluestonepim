package provider

import (
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Bluestone rate limits the API. With a few hundred resources and Terraform's
// parallel requests the default 5 attempts are not enough, so retry for a few
// minutes in total before giving up.
const (
	retryMax     = 10
	retryWaitMin = 1 * time.Second
	retryWaitMax = 60 * time.Second
)

// newRetryableClient returns a client that retries on rate limiting (429),
// server errors and connection errors. A `Retry-After` header is honoured,
// otherwise it backs off exponentially with jitter, so parallel requests that
// were rate limited together don't all retry at the same moment.
func newRetryableClient(max int, waitMin, waitMax time.Duration) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = max
	client.RetryWaitMin = waitMin
	client.RetryWaitMax = waitMax
	client.Backoff = backoffWithJitter
	return client
}

func backoffWithJitter(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleep := retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
	if resp != nil && resp.Header.Get("Retry-After") != "" {
		return sleep
	}
	// Up to 50% jitter, without going over the maximum
	sleep += time.Duration(rand.Int64N(int64(sleep)/2 + 1))
	return min64(sleep, max)
}

func min64(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
