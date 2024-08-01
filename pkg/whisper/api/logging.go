package api

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	req int32
)

func wrapLogging(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := nextReq()
		delta := time.Now()
		log.Printf("R%d %s %s", req, r.Method, r.URL)
		fn(w, r)
		log.Printf("R%d Took %v", req, time.Since(delta).Truncate(time.Millisecond))
	}
}

func nextReq() int32 {
	return atomic.AddInt32(&req, 1)
}
