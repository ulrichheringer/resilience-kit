# resilience-kit — Go HTTP Resilience Middleware

[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

HTTP middleware for Go servers. It includes a circuit breaker for failing upstreams and a concurrency limit that sheds excess requests.

Install with go get github.com/ulrichheringer/resilience-kit/resilience.
Run the example server with go run ./cmd/server, then try /work and /work?fail=1.
