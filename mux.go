package main

import (
	"fmt"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// If WriteHeader was not called explicitly, default status is 200
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// wrap the real ResponseWriter so we can inspect status & size
		rw := &responseWriter{ResponseWriter: w}

		// forward to the real mux (or next handler)
		next.ServeHTTP(rw, r)

		// if handler never wrote a header, assume 200
		if rw.status == 0 {
			rw.status = http.StatusOK
		}

		comp := w.Header().Get("Content-Encoding")
		if comp == "" {
			comp = "none"
		}
		fmt.Printf("[REQ] %s | %18s | %6s | %4s | %3d | %16s | %7s | %q\n",
			time.Now().Format("02/01/2006 - 03:04:05 PM"),
			r.RemoteAddr,
			r.Method,
			comp,
			rw.status,
			time.Since(start),
			byteCountIEC(rw.size),
			r.RequestURI,
		)
		// Log: remote, method, path, status, size, duration
		// fmt.Printf("[req] %s - %s %d %dB in %s -> %s \n",
		// 	r.RemoteAddr, r.Method, rw.status, rw.size, time.Since(start), r.RequestURI)
	})
}

func byteCountIEC[T int | int64](c T) string {
	b := int64(c)
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
