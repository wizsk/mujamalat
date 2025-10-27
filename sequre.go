package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cookieName      = "mujamalat_ss_id"
	sessionFileName = "sessions.txt" // plain text
	sessionExpiry   = 30 * 24 * time.Hour
)

// ===== GLOBAL STORE =====
type sequreSession struct {
	pass string
	file string
	sync.RWMutex
	t templateWraper
	m map[string]time.Time
}

var (
	sessionStore *sequreSession
)

// generateSessionID creates a random + hashed session token.
func generateSessionID() string {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random bytes: " + err.Error())
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// saveSessions writes valid (non-expired) sessions to plain text file
func saveSessions() {
	sessionStore.Lock()
	defer sessionStore.Unlock()

	var lines []string
	now := time.Now()

	for id, exp := range sessionStore.m {
		if now.Before(exp) {
			lines = append(lines, fmt.Sprintf("%s|%d", id, exp.Unix()))
		}
	}

	tmpFile := sessionStore.file + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		fmt.Println("Error saving sessions:", err)
		return
	}
	os.Rename(tmpFile, sessionStore.file)
}

// loadSessions loads sessions from plain text file
func loadSessions(dir, pass string, t templateWraper) {
	if dir == "" || pass == "" {
		lg.Fatal("either the root dir or pass is empty")
	}

	sessionStore = &sequreSession{
		file: filepath.Join(dir, sessionFileName),
		pass: pass,
		m:    make(map[string]time.Time),
		t:    t,
	}

	data, err := os.ReadFile(sessionStore.file)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Println("Error loading sessions:", err)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	sessionStore.Lock()
	defer sessionStore.Unlock()

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		expUnix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		if time.Now().Unix() < expUnix {
			sessionStore.m[parts[0]] = time.Unix(expUnix, 0)
		}
	}
	fmt.Printf("INFO: Loaded %d active sessions\n", len(sessionStore.m))
}

// cleanupExpired removes old sessions from memory and file
func cleanupExpired() {
	sessionStore.Lock()
	now := time.Now()
	count := 0
	for id, exp := range sessionStore.m {
		if now.After(exp) {
			delete(sessionStore.m, id)
			count++
		}
	}
	sessionStore.Unlock()
	if count > 0 {
		saveSessions()
	}
}

// background cleaner
func startCleanupTicker() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cleanupExpired()
		}
	}()
}

// ===== SESSION MANAGEMENT =====

func setSession(w http.ResponseWriter) {
	id := generateSessionID()
	expiry := time.Now().Add(sessionExpiry)

	sessionStore.Lock()
	sessionStore.m[id] = expiry
	sessionStore.Unlock()
	saveSessions()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // true in HTTPS
		Expires:  expiry,
	})
}

func isAuthenticated(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}

	sessionStore.RLock()
	expiry, ok := sessionStore.m[c.Value]
	sessionStore.RUnlock()

	if !ok || time.Now().After(expiry) {
		return false
	}
	return true
}

func logout(sessionID string) {
	sessionStore.Lock()
	delete(sessionStore.m, sessionID)
	sessionStore.Unlock()
	saveSessions()
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := sessionStore.t.ExecuteTemplate(w, loginTemplateName, nil); debug && err != nil {
			lg.Println(err)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		pass := r.FormValue("password")
		if subtle.ConstantTimeCompare([]byte(pass), []byte(sessionStore.pass)) != 1 {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
		setSession(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	c, err := r.Cookie(cookieName)
	if err == nil {
		logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func sequreMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/pub/") || r.URL.Path == "/auth" || isAuthenticated(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
		}
	})
}
