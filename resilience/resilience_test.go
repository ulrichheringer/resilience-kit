package resilience

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCircuitOpensAndAllowsOneProbe(t *testing.T) {
	b := NewCircuitBreaker(1, time.Millisecond)
	done, ok := b.Allow()
	if !ok {
		t.Fatal("first request should pass")
	}
	done(false)
	if _, ok := b.Allow(); ok {
		t.Fatal("open circuit accepted a request")
	}
	time.Sleep(2 * time.Millisecond)
	probe, ok := b.Allow()
	if !ok {
		t.Fatal("half-open probe should pass")
	}
	if _, ok := b.Allow(); ok {
		t.Fatal("more than one probe was accepted")
	}
	probe(true)
	if _, ok := b.Allow(); !ok {
		t.Fatal("closed circuit rejected a request")
	}
}

func TestMiddlewareReturns503WhenCircuitIsOpen(t *testing.T) {
	b := NewCircuitBreaker(1, time.Hour)
	handler := b.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "bad", 500) }))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 {
		t.Fatalf("status=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 503 {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestLoadShedderRejectsConcurrentRequest(t *testing.T) {
	s := NewLoadShedder(1)
	s.slots <- struct{}{}
	rec := httptest.NewRecorder()
	s.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 503 {
		t.Fatalf("status=%d", rec.Code)
	}
	<-s.slots
}
