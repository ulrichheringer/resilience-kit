package resilience

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type breakerState uint8

const (
	closed breakerState = iota
	open
	halfOpen
)

type CircuitBreaker struct {
	mu              sync.Mutex
	threshold       int
	resetAfter      time.Duration
	failures        int
	state           breakerState
	openedAt        time.Time
	probeInProgress bool
}

func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1
	}
	if resetAfter <= 0 {
		resetAfter = time.Second
	}
	return &CircuitBreaker{threshold: threshold, resetAfter: resetAfter}
}

func (b *CircuitBreaker) Allow() (finish func(success bool), allowed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == open {
		if time.Since(b.openedAt) < b.resetAfter {
			return nil, false
		}
		b.state = halfOpen
	}
	if b.state == halfOpen {
		if b.probeInProgress {
			return nil, false
		}
		b.probeInProgress = true
	}
	var once sync.Once
	return func(success bool) { once.Do(func() { b.complete(success) }) }, true
}

func (b *CircuitBreaker) complete(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == halfOpen {
		b.probeInProgress = false
		if success {
			b.state = closed
			b.failures = 0
		} else {
			b.state = open
			b.openedAt = time.Now()
		}
		return
	}
	if success {
		b.failures = 0
		return
	}
	b.failures++
	if b.failures >= b.threshold {
		b.state = open
		b.openedAt = time.Now()
	}
}

func (b *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finish, allowed := b.Allow()
		if !allowed {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "circuit open", http.StatusServiceUnavailable)
			return
		}
		recorder := &statusWriter{ResponseWriter: w}
		defer func() { finish(recorder.status < 500) }()
		next.ServeHTTP(recorder, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type LoadShedder struct{ slots chan struct{} }

func NewLoadShedder(maxConcurrent int) *LoadShedder {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &LoadShedder{slots: make(chan struct{}, maxConcurrent)}
}

func (s *LoadShedder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case s.slots <- struct{}{}:
			defer func() { <-s.slots }()
			next.ServeHTTP(w, r)
		default:
			w.Header().Set("Retry-After", "1")
			w.Header().Set("X-Load-Shed", strconv.Itoa(cap(s.slots)))
			http.Error(w, "server busy", http.StatusServiceUnavailable)
		}
	})
}
