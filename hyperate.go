package hyperate

import (
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/time/rate"
)

var _ http.RoundTripper = (*RateLimitRoundTripper)(nil)

// RateLimitRoundTripper is an http.RoundTripper that rate limits requests.
type RateLimitRoundTripper struct {
	trans http.RoundTripper
	lim   *rate.Limiter

	// onResp is a function that is called after an HTTP response is received.
	// It allows the caller to inspect the response, e.g. to check for rate
	// limit headers. If onResp is nil, the response is returned as-is.
	onResp OnRespFunc
}

type OnRespFunc func(*http.Response, error) (*http.Response, error)

type RateLimitRoundTripperOption func(*RateLimitRoundTripper)

// New returns a a new RateLimitRoundtripper.
func New(roudtripper http.RoundTripper, limiter *rate.Limiter, opts ...RateLimitRoundTripperOption) *RateLimitRoundTripper {
	rt := &RateLimitRoundTripper{
		trans: roudtripper,
		lim:   limiter,
	}

	for _, option := range opts {
		option(rt)
	}

	return rt
}

func WithOnRespFunc(fn OnRespFunc) RateLimitRoundTripperOption {
	return func(rt *RateLimitRoundTripper) {
		rt.onResp = fn
	}
}

// RoundTrip implements http.RoundTripper.
//
// It blocks, waiting until the rate limiter permits the request to be made,
// then delegates to the underlying RoundTripper to make the request. If the
// request's context is canceled or its deadline is exceeded while waiting, an
// error is returned.
func (rt *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := rt.lim.Wait(ctx); err != nil {
		return nil, err
	}

	res, err := rt.trans.RoundTrip(req)

	if rt.onResp != nil {
		return rt.onResp(res, err)
	}

	return res, err
}

// WithRateLimitHeaderCheck returns a RateLimitRoundTripperOption that checks
// the response headers for rate limit information.
//
// See: https://www.ietf.org/archive/id/draft-polli-ratelimit-headers-02.html
func WithRateLimitHeaderCheck() RateLimitRoundTripperOption {
	return func(rt *RateLimitRoundTripper) {
		// Add middleware to check and use the response headers for rate
		// limiting.
		rt.onResp = func(res *http.Response, respErr error) (*http.Response, error) {
			if respErr != nil || rt.lim == nil {
				return res, respErr
			}

			remainingHeader := res.Header.Get("RateLimit-Remaining")
			resetHeader := res.Header.Get("RateLimit-Reset")

			if remainingHeader != "" && resetHeader != "" {
				// The amount of requests remaining until the rate limit is
				// reached.
				remaining, err := strconv.Atoi(remainingHeader)
				if err != nil {
					return nil, fmt.Errorf("hyperate: failed to parse rate limit remaining header: %w", err)
				}

				// The amount of seconds until the rate limit is reset.
				reset, err := strconv.Atoi(resetHeader)
				if err != nil {
					return nil, fmt.Errorf("hyperate: failed to parse rate limit reset header: %w", err)
				}

				// Evenly spread the remaining requests over the remaining time
				// until the rate limit window is reset.
				if reset > 0 {
					rt.lim.SetLimit(rate.Limit(remaining / reset))
				}
			}

			return res, respErr
		}
	}
}
