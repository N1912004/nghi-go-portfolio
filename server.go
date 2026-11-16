package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"html"
	"html/template"
	"io/fs"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Config struct {
	PublicURL  string
	Production bool
}
type GameScore struct {
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

type App struct {
	profile     Profile
	store       Store
	cfg         Config
	template    *template.Template
	limiter     *Limiter
	started     time.Time
	logger      *slog.Logger
	gameMu      sync.RWMutex
	scores      []GameScore
	gameLimiter *Limiter
}
type View struct {
	Profile   Profile
	Projects  []Project
	CSRF      string
	Sent      bool
	Error     string
	Mode      string
	GoVersion string
	Year      int
	PublicURL string
}

func NewApp(p Profile, store Store, cfg Config, logger *slog.Logger) (*App, error) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{"addOne": func(i int) int { return i + 1 }}).ParseFS(assets, "web/index.html")
	if err != nil {
		return nil, err
	}
	initialScores := []GameScore{
		{Name: "GopherPro", Score: 2450, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{Name: "BackendNinja", Score: 1820, CreatedAt: time.Now().Add(-5 * time.Hour)},
		{Name: "ConcurrencyMaster", Score: 1400, CreatedAt: time.Now().Add(-12 * time.Hour)},
		{Name: "RedisCacheKing", Score: 950, CreatedAt: time.Now().Add(-24 * time.Hour)},
		{Name: "GoRookie", Score: 500, CreatedAt: time.Now().Add(-48 * time.Hour)},
	}
	return &App{
		profile:     p,
		store:       store,
		cfg:         cfg,
		template:    tmpl,
		limiter:     NewLimiter(5, 10*time.Minute),
		started:     time.Now(),
		logger:      logger,
		scores:      initialScores,
		gameLimiter: NewLimiter(10, 1*time.Minute),
	}, nil
}
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	static, _ := fs.Sub(assets, "web/assets")
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(static))))
	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /api/projects", a.projects)
	mux.HandleFunc("GET /api/profile", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, a.profile) })
	mux.HandleFunc("GET /api/game/leaderboard", a.getLeaderboard)
	mux.HandleFunc("POST /api/game/score", a.postScore)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", a.ready)
	mux.HandleFunc("POST /contact", a.contact)
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /api/\nDisallow: /readyz\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	return a.secure(mux)
}
func (a *App) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		if a.cfg.Production {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}
		defer func() {
			if recover() != nil {
				a.logger.Error("request panic", "path", r.URL.Path)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (a *App) cookieName() string {
	if a.cfg.Production {
		return "__Host-portfolio_csrf"
	}
	return "portfolio_csrf"
}
func (a *App) issueToken(w http.ResponseWriter) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: a.cookieName(), Value: token, Path: "/", MaxAge: 3600, HttpOnly: true, Secure: a.cfg.Production, SameSite: http.SameSiteStrictMode})
	return token, nil
}
func (a *App) home(w http.ResponseWriter, r *http.Request) { a.render(w, r, http.StatusOK, "") }
func (a *App) render(w http.ResponseWriter, r *http.Request, status int, message string) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	projects, err := a.store.Projects(ctx)
	if err != nil {
		a.logger.Error("projects unavailable")
		http.Error(w, "Portfolio temporarily unavailable. Please try again.", 503)
		return
	}
	token, err := a.issueToken(w)
	if err != nil {
		http.Error(w, "Unable to initialize form", 500)
		return
	}
	var b bytes.Buffer
	view := View{Profile: a.profile, Projects: projects, CSRF: token, Sent: r.URL.Query().Get("sent") == "1", Error: message, Mode: a.store.Mode(), GoVersion: runtime.Version(), Year: time.Now().Year(), PublicURL: a.cfg.PublicURL}
	if err = a.template.Execute(&b, view); err != nil {
		a.logger.Error("template render failed")
		http.Error(w, "Unable to render portfolio", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(b.Bytes())
}
func (a *App) projects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	p, err := a.store.Projects(ctx)
	if err != nil {
		writeJSON(w, 503, map[string]string{"error": "projects temporarily unavailable"})
		return
	}
	writeJSON(w, 200, p)
}
func (a *App) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	a.gameMu.RLock()
	defer a.gameMu.RUnlock()

	limit := 10
	if len(a.scores) < limit {
		limit = len(a.scores)
	}
	res := make([]GameScore, limit)
	copy(res, a.scores[:limit])
	writeJSON(w, http.StatusOK, res)
}

type postScoreRequest struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func (a *App) postScore(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if !a.gameLimiter.Allow(host, time.Now()) {
		w.Header().Set("Retry-After", "60")
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Too many requests. Please wait a minute."})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var req postScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		return
	}

	name := strings.TrimSpace(html.EscapeString(req.Name))
	nameLen := utf8.RuneCountInString(name)
	if nameLen < 2 || nameLen > 25 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Player name must be between 2 and 25 characters"})
		return
	}
	if strings.ContainsAny(name, "\r\n\x00") {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Invalid player name characters"})
		return
	}
	if req.Score <= 0 || req.Score > 100000 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Score must be between 1 and 100,000"})
		return
	}

	entry := GameScore{
		Name:      name,
		Score:     req.Score,
		CreatedAt: time.Now(),
	}

	a.gameMu.Lock()
	a.scores = append(a.scores, entry)
	sort.SliceStable(a.scores, func(i, j int) bool {
		return a.scores[i].Score > a.scores[j].Score
	})
	if len(a.scores) > 50 {
		a.scores = a.scores[:50]
	}
	limit := 10
	if len(a.scores) < limit {
		limit = len(a.scores)
	}
	topList := make([]GameScore, limit)
	copy(topList, a.scores[:limit])
	a.gameMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "Score recorded successfully",
		"leaderboard": topList,
	})
}
func (a *App) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.store.Ping(ctx); err != nil {
		writeJSON(w, 503, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, 200, map[string]any{"status": "ready", "language": "Go", "runtime": runtime.Version(), "storage": a.store.Mode(), "uptime_seconds": int64(time.Since(a.started).Seconds())})
}
func (a *App) contact(w http.ResponseWriter, r *http.Request) {
	// Do not trust X-Forwarded-For from arbitrary clients. Behind a reverse proxy,
	// the proxy address may share this quota; documented single-instance limiter.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if !a.limiter.Allow(host, time.Now()) {
		w.Header().Set("Retry-After", "600")
		a.render(w, r, 429, "Too many attempts. Please wait ten minutes or email me directly.")
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/x-www-form-urlencoded" {
		http.Error(w, "Expected form data", 415)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != a.cfg.PublicURL {
		http.Error(w, "Invalid request origin", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Form is too large or invalid", 400)
		return
	}
	cookie, err := r.Cookie(a.cookieName())
	token := r.PostForm.Get("csrf_token")
	if err != nil || len(token) != 64 || len(cookie.Value) != 64 || subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) != 1 {
		a.render(w, r, 403, "The form expired. Please submit it again.")
		return
	}
	if r.PostForm.Get("website") != "" {
		http.Error(w, "Invalid submission", 400)
		return
	}
	c := Contact{Name: strings.TrimSpace(r.PostForm.Get("name")), Email: strings.TrimSpace(r.PostForm.Get("email")), Message: strings.TrimSpace(r.PostForm.Get("message"))}
	if reason := validateContact(c); reason != "" {
		a.render(w, r, 422, reason)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err = a.store.SaveContact(ctx, c); err != nil {
		a.logger.Error("contact storage failed")
		a.render(w, r, 503, "Your message could not be saved. Please email me directly.")
		return
	}
	// Never log message text/email. No outgoing email is claimed or performed.
	http.Redirect(w, r, "/?sent=1#contact", http.StatusSeeOther)
}
func validateContact(c Contact) string {
	n := utf8.RuneCountInString(c.Name)
	if n < 2 || n > 80 {
		return "Please enter a name between 2 and 80 characters."
	}
	if strings.ContainsAny(c.Name, "\r\n\x00") {
		return "Please enter a valid name."
	}
	addr, err := mail.ParseAddress(c.Email)
	if err != nil || addr.Address != c.Email || len(c.Email) > 254 || strings.ContainsAny(c.Email, "\r\n") {
		return "Please enter a valid email address."
	}
	n = utf8.RuneCountInString(c.Message)
	if n < 20 || n > 2000 || strings.ContainsRune(c.Message, '\x00') {
		return "Your message must contain 20–2,000 characters."
	}
	return ""
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type limitEntry struct {
	Count int
	Start time.Time
}
type Limiter struct {
	mu      sync.Mutex
	entries map[string]limitEntry
	limit   int
	window  time.Duration
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{entries: make(map[string]limitEntry), limit: limit, window: window}
}
func (l *Limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Bounded memory, with opportunistic removal of expired clients.
	if len(l.entries) >= 4096 {
		for k, v := range l.entries {
			if now.Sub(v.Start) >= l.window {
				delete(l.entries, k)
			}
		}
	}
	e, exists := l.entries[key]
	if !exists && len(l.entries) >= 4096 {
		return false
	}
	if !exists || now.Sub(e.Start) >= l.window {
		e = limitEntry{Start: now}
	}
	if e.Count >= l.limit {
		return false
	}
	e.Count++
	l.entries[key] = e
	return true
}
func validatePublicURL(raw string, production bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return &url.Error{Op: "configure", URL: "PUBLIC_URL", Err: errInvalidURL{}}
	}
	if production && u.Scheme != "https" {
		return errInvalidURL{}
	}
	return nil
}

type errInvalidURL struct{}

func (errInvalidURL) Error() string {
	return "PUBLIC_URL must be an origin such as https://portfolio.example.com (HTTPS required in production)"
}
