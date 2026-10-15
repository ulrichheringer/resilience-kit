package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ulrichheringer/resilience-kit/resilience"
)

func main() {
	breaker := resilience.NewCircuitBreaker(3, 5*time.Second)
	shedder := resilience.NewLoadShedder(20)
	work := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(250 * time.Millisecond)
		if r.URL.Query().Get("fail") == "1" {
			http.Error(w, "simulated upstream error", http.StatusBadGateway)
			return
		}
		fmt.Fprintln(w, "work completed")
	})
	http.Handle("/work", shedder.Middleware(breaker.Middleware(work)))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
