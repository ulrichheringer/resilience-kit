# resilience-kit

HTTP middleware for Go servers. It includes a circuit breaker for failing upstreams and a concurrency limit that sheds excess requests.

Install with go get github.com/ulrichheringer/resilience-kit/resilience.
Run the example server with go run ./cmd/server, then try /work and /work?fail=1.
