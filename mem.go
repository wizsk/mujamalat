// ai made
package main

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// MemoryMonitor stores historical memory data
type MemoryMonitor struct {
	mu      sync.RWMutex
	history []MemorySnapshot
	maxAge  time.Duration
}

// MemorySnapshot represents memory stats at a point in time
type MemorySnapshot struct {
	Timestamp  int64   `json:"timestamp"`  // Unix timestamp in milliseconds
	Alloc      uint64  `json:"alloc"`      // bytes
	TotalAlloc uint64  `json:"totalAlloc"` // bytes
	Sys        uint64  `json:"sys"`        // bytes
	HeapAlloc  uint64  `json:"heapAlloc"`  // bytes
	HeapSys    uint64  `json:"heapSys"`    // bytes
	HeapIdle   uint64  `json:"heapIdle"`   // bytes
	HeapInuse  uint64  `json:"heapInuse"`  // bytes
	NumGC      uint32  `json:"numGC"`
	GCPauseMs  float64 `json:"gcPauseMs"` // milliseconds
}

var monitor *MemoryMonitor

func doMeminit() {
	monitor = &MemoryMonitor{
		history: make([]MemorySnapshot, 0, 3600), // Pre-allocate for 1 hour at 1s interval
		maxAge:  time.Hour,
	}

	// Start background collector
	go monitor.collect()
}

// collect gathers memory stats every second
func (m *MemoryMonitor) collect() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		snapshot := m.captureSnapshot()
		m.addSnapshot(snapshot)
	}
}

// captureSnapshot captures current memory stats
func (m *MemoryMonitor) captureSnapshot() MemorySnapshot {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return MemorySnapshot{
		Timestamp:  time.Now().UnixMilli(),
		Alloc:      ms.Alloc,
		TotalAlloc: ms.TotalAlloc,
		Sys:        ms.Sys,
		HeapAlloc:  ms.HeapAlloc,
		HeapSys:    ms.HeapSys,
		HeapIdle:   ms.HeapIdle,
		HeapInuse:  ms.HeapInuse,
		NumGC:      ms.NumGC,
		GCPauseMs:  float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1e6,
	}
}

// addSnapshot adds a snapshot and removes old ones
func (m *MemoryMonitor) addSnapshot(snapshot MemorySnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, snapshot)

	// Remove snapshots older than maxAge
	cutoff := time.Now().Add(-m.maxAge).UnixMilli()
	for i, snap := range m.history {
		if snap.Timestamp >= cutoff {
			m.history = m.history[i:]
			break
		}
	}
}

// getHistory returns snapshots within the specified duration
func (m *MemoryMonitor) getHistory(duration time.Duration) []MemorySnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().Add(-duration).UnixMilli()
	result := make([]MemorySnapshot, 0)

	for _, snap := range m.history {
		if snap.Timestamp >= cutoff {
			result = append(result, snap)
		}
	}

	return result
}

// MemAPIHandler handles API requests for memory data
func MemAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Parse duration from query parameter (default 1 hour)
	durationParam := r.URL.Query().Get("duration")
	duration := time.Hour

	if durationParam != "" {
		if d, err := time.ParseDuration(durationParam); err == nil {
			duration = d
		}
	}

	history := monitor.getHistory(duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]MemorySnapshot{
		"data": history,
	})
}

// Setup registers the handlers
func SetupMemoryHandlers(mux *http.ServeMux) {
	doMeminit()
	mux.HandleFunc("/mem", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api") != "" {
			MemAPIHandler(w, r)
		}
	})
}
