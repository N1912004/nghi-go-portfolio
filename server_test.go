package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func testApp(t *testing.T) (*App, *MemoryStore) {
	t.Helper()
	p, err := loadProfile()
	if err != nil {
		t.Fatal(err)
	}
	s := NewMemoryStore(p.Projects)
	a, err := NewApp(p, s, Config{PublicURL: "http://example.com"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return a, s
}
func getForm(t *testing.T, a *App) (*http.Cookie, string) {
	t.Helper()
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "http://example.com/", nil))
	if w.Code != 200 {
		t.Fatalf("home=%d: %s", w.Code, w.Body.String())
	}
	var c *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == a.cookieName() {
			c = cookie
		}
	}
	if c == nil || len(c.Value) != 64 {
		t.Fatal("missing CSRF cookie")
	}
	return c, c.Value
}
func submit(a *App, v url.Values, c *http.Cookie, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://example.com/contact", strings.NewReader(v.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if c != nil {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	return w
}
func validForm(token string) url.Values {
	return url.Values{"csrf_token": {token}, "name": {"Nguyen Nghi"}, "email": {"recruiter@example.com"}, "message": {"Hello, I would like to discuss a backend role."}}
}
func TestHome(t *testing.T) {
	a, _ := testApp(t)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, text := range []string{"Nguyen Quoc Nghi", "Go AI Job Gateway", "csrf_token", "Development mode"} {
		if !strings.Contains(w.Body.String(), text) {
			t.Errorf("missing %q", text)
		}
	}
}
func TestSecurityHeaders(t *testing.T) {
	a, _ := testApp(t)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	for _, h := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if w.Header().Get(h) == "" {
			t.Errorf("missing %s", h)
		}
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("CSRF page must not be cached")
	}
}
func TestAssets(t *testing.T) {
	a, _ := testApp(t)
	for _, path := range []string{"/assets/style.css", "/assets/app.js", "/assets/portrait.jpg", "/assets/resume.pdf", "/assets/favicon.svg"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			a.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code != 200 || w.Body.Len() == 0 {
				t.Fatal(path, w.Code)
			}
		})
	}
}
func TestProjectsAPI(t *testing.T) {
	a, _ := testApp(t)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/projects", nil))
	var p []Project
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil || len(p) != 5 {
		t.Fatal(err, w.Body.String())
	}
	if p[0].Category != "Personal / Go" {
		t.Fatal("project scope lost")
	}
}
func TestLivenessAndReadiness(t *testing.T) {
	a, _ := testApp(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatal(path, w.Code)
		}
	}
}
func TestContactSaved(t *testing.T) {
	a, s := testApp(t)
	c, token := getForm(t, a)
	w := submit(a, validForm(token), c, "http://example.com")
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w.Header().Get("Location") != "/?sent=1#contact" || len(s.contacts) != 1 {
		t.Fatal("contact not saved")
	}
}
func TestMissingCSRF(t *testing.T) {
	a, s := testApp(t)
	w := submit(a, validForm(""), nil, "")
	if w.Code != 403 || len(s.contacts) != 0 {
		t.Fatal(w.Code)
	}
}
func TestForgedCSRF(t *testing.T) {
	a, s := testApp(t)
	c, _ := getForm(t, a)
	w := submit(a, validForm(strings.Repeat("0", 64)), c, "")
	if w.Code != 403 || len(s.contacts) != 0 {
		t.Fatal(w.Code)
	}
}
func TestCrossOriginRejected(t *testing.T) {
	a, s := testApp(t)
	c, token := getForm(t, a)
	w := submit(a, validForm(token), c, "https://evil.example")
	if w.Code != 403 || len(s.contacts) != 0 {
		t.Fatal(w.Code)
	}
}
func TestHoneypot(t *testing.T) {
	a, s := testApp(t)
	c, token := getForm(t, a)
	v := validForm(token)
	v.Set("website", "spam")
	w := submit(a, v, c, "")
	if w.Code != 400 || len(s.contacts) != 0 {
		t.Fatal(w.Code)
	}
}
func TestContactValidation(t *testing.T) {
	cases := []struct {
		name string
		edit func(url.Values)
	}{
		{"short name", func(v url.Values) { v.Set("name", "A") }},
		{"long name", func(v url.Values) { v.Set("name", strings.Repeat("N", 81)) }},
		{"invalid email", func(v url.Values) { v.Set("email", "broken") }},
		{"email display name", func(v url.Values) { v.Set("email", "Nghi <nghi@example.com>") }},
		{"short message", func(v url.Values) { v.Set("message", "Hello") }},
		{"long message", func(v url.Values) { v.Set("message", strings.Repeat("x", 2001)) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, s := testApp(t)
			c, token := getForm(t, a)
			v := validForm(token)
			tc.edit(v)
			w := submit(a, v, c, "")
			if w.Code != 422 || len(s.contacts) != 0 {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
}
func TestUnicodeValidation(t *testing.T) {
	c := Contact{Name: "Nguyễn Quốc Nghi", Email: "nghi@example.com", Message: strings.Repeat("ệ", 30)}
	if reason := validateContact(c); reason != "" {
		t.Fatal(reason)
	}
}
func TestOversizedRequest(t *testing.T) {
	a, _ := testApp(t)
	c, token := getForm(t, a)
	v := validForm(token)
	v.Set("message", strings.Repeat("a", 20<<10))
	w := submit(a, v, c, "")
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestContactRateLimit(t *testing.T) {
	a, _ := testApp(t)
	c, token := getForm(t, a)
	for i := 0; i < 5; i++ {
		if w := submit(a, validForm(token), c, ""); w.Code != 303 {
			t.Fatal(i, w.Code)
		}
	}
	w := submit(a, validForm(token), c, "")
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Code)
	}
}
func TestMethodRestrictions(t *testing.T) {
	a, _ := testApp(t)
	for _, tc := range []struct{ method, path string }{{"POST", "/api/projects"}, {"POST", "/healthz"}, {"GET", "/contact"}} {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code == 200 {
			t.Fatal(tc, "unexpected success")
		}
	}
}
func TestNotFound(t *testing.T) {
	a, _ := testApp(t)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/missing", nil))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
func TestProductionCookie(t *testing.T) {
	a, _ := testApp(t)
	a.cfg = Config{PublicURL: "https://example.com", Production: true}
	c, _ := getForm(t, a)
	if c.Name != "__Host-portfolio_csrf" || !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unsafe cookie: %+v", c)
	}
}
func TestTemplateEscapesContent(t *testing.T) {
	a, _ := testApp(t)
	a.profile.Name = "<script>alert(1)</script>"
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if strings.Contains(w.Body.String(), a.profile.Name) {
		t.Fatal("unescaped profile content")
	}
}
func TestLimiterConcurrent(t *testing.T) {
	l := NewLimiter(5, time.Minute)
	now := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	n := 0
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("client", now) {
				mu.Lock()
				n++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if n != 5 {
		t.Fatal(n)
	}
	if !l.Allow("client", now.Add(time.Minute)) {
		t.Fatal("quota did not reset")
	}
}
func TestMemoryIsolationAndCancellation(t *testing.T) {
	a, s := testApp(t)
	p, _ := s.Projects(context.Background())
	p[0].Stack[0] = "MUTATED"
	p2, _ := s.Projects(context.Background())
	if p2[0].Stack[0] == "MUTATED" {
		t.Fatal("mutable state leaked")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if s.Ping(ctx) == nil {
		t.Fatal("cancellation ignored")
	}
	_ = a
}
func TestPublicURLValidation(t *testing.T) {
	for _, tc := range []struct {
		v              string
		production, ok bool
	}{{"http://localhost:8080", false, true}, {"https://example.com", true, true}, {"http://example.com", true, false}, {"https://example.com/path", true, false}, {"https://example.com?x=1", true, false}, {"https://user:pass@example.com", true, false}, {"broken", false, false}} {
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			if (validatePublicURL(tc.v, tc.production) == nil) != tc.ok {
				t.Fatal(tc)
			}
		})
	}
}

type failingStore struct{ Store }

func (failingStore) Ping(context.Context) error                 { return errors.New("offline") }
func (failingStore) SaveContact(context.Context, Contact) error { return errors.New("offline") }
func TestReadinessFailsClosed(t *testing.T) {
	a, s := testApp(t)
	a.store = failingStore{s}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}
func TestContactFailureNotSuccess(t *testing.T) {
	a, s := testApp(t)
	a.store = failingStore{s}
	c, token := getForm(t, a)
	w := submit(a, validForm(token), c, "")
	if w.Code != 503 || strings.Contains(w.Header().Get("Location"), "sent=1") {
		t.Fatal(w.Code)
	}
}

func TestGameEndpoints(t *testing.T) {
	a, _ := testApp(t)

	// Test GET Leaderboard
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/game/leaderboard", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var scores []GameScore
	if err := json.Unmarshal(w.Body.Bytes(), &scores); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("expected scores, got empty")
	}

	// Test valid POST score
	postData := `{"name": "Champion", "score": 3500}`
	req := httptest.NewRequest("POST", "/api/game/score", strings.NewReader(postData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Message     string      `json:"message"`
		Leaderboard []GameScore `json:"leaderboard"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(resp.Leaderboard) == 0 || resp.Leaderboard[0].Name != "Champion" || resp.Leaderboard[0].Score != 3500 {
		t.Fatalf("expected Champion at rank 1, got %+v", resp.Leaderboard[0])
	}

	// Test invalid score (negative or zero)
	req = httptest.NewRequest("POST", "/api/game/score", strings.NewReader(`{"name": "Noob", "score": -5}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}

	// Test invalid name length
	req = httptest.NewRequest("POST", "/api/game/score", strings.NewReader(`{"name": "X", "score": 100}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}

	// Test name sanitization (HTML escaping)
	req = httptest.NewRequest("POST", "/api/game/score", strings.NewReader(`{"name": "<b>Hacker</b>", "score": 2100}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var sanitizeResp struct {
		Leaderboard []GameScore `json:"leaderboard"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &sanitizeResp)
	found := false
	for _, s := range sanitizeResp.Leaderboard {
		if strings.Contains(s.Name, "Hacker") {
			found = true
			if strings.Contains(s.Name, "<script>") || strings.Contains(s.Name, "<b>") {
				t.Fatalf("name was not properly escaped: %s", s.Name)
			}
		}
	}
	if !found {
		t.Fatal("expected sanitized player in leaderboard")
	}

	// Test boundary score
	req = httptest.NewRequest("POST", "/api/game/score", strings.NewReader(`{"name": "OverLimit", "score": 100001}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for score > 100000, got %d", w.Code)
	}
}
