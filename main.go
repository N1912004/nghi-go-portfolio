package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
func run(logger *slog.Logger) error {
	loadDotEnv()
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("PORT must be between 1 and 65535")
	}
	production := os.Getenv("APP_ENV") == "production"
	publicURL := strings.TrimRight(os.Getenv("PUBLIC_URL"), "/")
	if publicURL == "" {
		publicURL = strings.TrimRight(os.Getenv("RENDER_EXTERNAL_URL"), "/")
	}
	if publicURL == "" {
		if production {
			return errors.New("set PUBLIC_URL in production")
		}
		publicURL = "http://localhost:" + port
	}
	if err := validatePublicURL(publicURL, production); err != nil {
		return err
	}
	profile, err := loadProfile()
	if err != nil {
		return err
	}
	var store Store
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		if production {
			return errors.New("DATABASE_URL is required in production; no silent memory fallback")
		}
		store = NewMemoryStore(profile.Projects)
		logger.Warn("explicit development memory mode: contact messages are temporary")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		store, err = NewPostgresStore(ctx, dsn, profile.Projects)
		cancel()
		if err != nil {
			return err
		}
	}
	defer store.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if pg, ok := store.(*PostgresStore); ok {
		cleanup := func() {
			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := pg.DeleteExpired(c); err != nil && ctx.Err() == nil {
				logger.Error("contact retention cleanup failed")
			}
		}
		cleanup()
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cleanup()
				}
			}
		}()
	}
	app, err := NewApp(profile, store, Config{PublicURL: publicURL, Production: production}, logger)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	candidatePorts := []string{port}
	if sp := os.Getenv("SERVER_PORT"); sp != "" && sp != port {
		candidatePorts = append(candidatePorts, sp)
	}
	for _, fallback := range []string{"8080", "3000"} {
		found := false
		for _, cp := range candidatePorts {
			if cp == fallback {
				found = true
				break
			}
		}
		if !found {
			candidatePorts = append(candidatePorts, fallback)
		}
	}

	var listeners []net.Listener
	for _, p := range candidatePorts {
		l, err := net.Listen("tcp", "0.0.0.0:"+p)
		if err == nil {
			listeners = append(listeners, l)
			logger.Info("listening", "address", l.Addr().String(), "storage", store.Mode(), "production", production)
		} else {
			logger.Warn("could not bind port", "port", p, "error", err)
		}
	}
	if len(listeners) == 0 {
		return errors.New("failed to bind to any port")
	}

	errorsCh := make(chan error, len(listeners))
	for _, l := range listeners {
		l := l
		go func() {
			if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errorsCh <- err
			}
		}()
	}
	select {
	case err := <-errorsCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			_ = srv.Close()
			return err
		}
		logger.Info("graceful shutdown complete")
		return nil
	}
}

func loadDotEnv() {
	paths := []string{".env", "/.env", "/home/container/.env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}
