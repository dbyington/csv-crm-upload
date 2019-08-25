package main

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
)

// This server guarantees 90% availability. Meaning 90% of your requests will get serviced. :-)
const passPercent = 90

type crm struct {
	mutex  sync.Mutex
	total  int64
	failed int64
}

func (c *crm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.total++
	statusCode := http.StatusOK
	if r.Method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	if !pass(c.total, c.failed, passPercent) {
		c.failed++
		log.Printf("failing request, %d failures", c.failed)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(""))
		return
	}
	w.WriteHeader(statusCode)
	w.Write([]byte(""))
}

func main() {
	s := &http.Server{Addr: ":8089"}
	http.Handle("/", &crm{})
	s.ListenAndServe()
}

func pass(t, f, p int64) bool {
	if t == 0 {
		return true
	}
	i := rand.Int63n(100)
	passRate := t - f/t
	if i > p {
		if passRate < p {
			return true
		}
		return false
	}
	return true
}
