package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"maps"
	"net/http"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

const (
	minSizeToCompress = 1024 // 1 KB
)

var (
	// Gzip writer pool
	gzipPool = sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression) // level 6
			return w
		},
	}

	// Brotli writer pool
	brPool = sync.Pool{
		New: func() any {
			return brotli.NewWriterLevel(nil, 5) // recommended for dynamic content
		},
	}

	// MIME types we allow to compress
	compressibleMIME = []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/xhtml+xml",
		"image/svg+xml",
	}
)

// Checks if MIME is compressible
func isCompressible(mimeType string) bool {
	for _, m := range compressibleMIME {
		if strings.HasPrefix(mimeType, m) {
			return true
		}
	}
	return false
}

// Wraps ResponseWriter to capture output
type captureWriter struct {
	http.ResponseWriter
	status int
	buf    *bytes.Buffer
	header http.Header
}

func newCaptureWriter(w http.ResponseWriter) *captureWriter {
	return &captureWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		buf:            getBuf(),
		header:         make(http.Header),
	}
}

func (cw *captureWriter) Header() http.Header {
	return cw.header
}

func (cw *captureWriter) WriteHeader(status int) {
	cw.status = status
}

func (cw *captureWriter) Write(b []byte) (int, error) {
	return cw.buf.Write(b)
}

// Middleware
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ae := r.Header.Get("Accept-Encoding")
		if !(strings.Contains(ae, "gzip") || strings.Contains(ae, "br")) {
			next.ServeHTTP(w, r) // no compression support
			return
		}

		cw := newCaptureWriter(w)
		defer putBuf(cw.buf)

		next.ServeHTTP(cw, r)

		maps.Copy(w.Header(), cw.header)

		// TODO: LOOK INTO IT
		// // Compute ETag
		// hash := sha1.Sum(cw.buf.Bytes())
		// etag := fmt.Sprintf(`"%x"`, hash[:])
		// w.Header().Set("ETag", etag)
		//
		// // 304 Not Modified
		// if match := r.Header.Get("If-None-Match"); match != "" {
		// 	if strings.Contains(match, etag) {
		// 		w.WriteHeader(http.StatusNotModified)
		// 		return
		// 	}
		// }

		// Determine MIME type
		contentType := cw.header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(cw.buf.Bytes())
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
		}

		if !isCompressible(contentType) ||
			cw.buf.Len() < minSizeToCompress {
			w.WriteHeader(cw.status)
			w.Write(cw.buf.Bytes())
			return
		}

		w.Header().Set("Vary", "Accept-Encoding")

		// Prefer Brotli
		if strings.Contains(ae, "br") {
			w.Header().Set("Content-Encoding", "br")

			bw := brPool.Get().(*brotli.Writer)
			bw.Reset(w)

			_, _ = io.Copy(bw, cw.buf)
			bw.Close()

			brPool.Put(bw)
			return
		}

		// Fallback to gzip
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)

		_, _ = io.Copy(gz, cw.buf)
		gz.Close()

		gzipPool.Put(gz)
	})
}
