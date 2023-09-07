# hyperate ⏱

**hyperate** provides a wrapper around a `http.RoundTripper` to help rate limit
outgoing HTTP requests.

## Installation

```shell
go get github.com/dstotijn/hyperate
```

## Usage

### Simple

```go
// Create an http.RoundTripper that rate limits outgoing HTTP requests to 10
// per second, allowing bursts of 5 requests.
rt := hyperate.New(http.DefaultTransport, rate.NewLimiter(10, 5))
client := &http.Client{Transport: rt}
```

### Based on HTTP response headers

Use HTTP response headers (`RateLimit-Remaining` and `RateLimit-Reset`)
([draft-polli-ratelimit-headers-02](https://www.ietf.org/archive/id/draft-polli-ratelimit-headers-02.html#name-ratelimit-limit))
to continously set rate limit.

```go
rt := hyperate.New(http.DefaultTransport, rate.NewLimiter(rate.Inf, 1),
    hyperate.WithRateLimitHeaderCheck(),
)
client := &http.Client{Transport: rt}
```

## License

[MIT](LICENSE)

---

© 2023 David Stotijn
