package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/config"
	"dinosaur/internal/server"
	"dinosaur/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
func run(logger *slog.Logger) error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	startup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := store.Open(startup, c.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	initialized, err := db.Initialized(startup)
	if err != nil {
		return err
	}
	if !initialized && c.SetupToken == "" {
		tokenPath := filepath.Join(c.DataDir, "setup-token")
		c.SetupToken, err = setupToken(tokenPath)
		if err != nil {
			return fmt.Errorf("initialize setup token: %w", err)
		}
		logger.Info("administrator setup required; read token file locally", "token_file", tokenPath)
	}
	handler, err := server.New(c, db, logger)
	if err != nil {
		return err
	}
	defer handler.Close()
	if err := handler.StartScanning(); err != nil {
		return fmt.Errorf("start media scanner: %w", err)
	}
	srv := &http.Server{Addr: c.Addr, Handler: handler.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: c.RequestTimeout, WriteTimeout: c.RequestTimeout + 5*time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 * 1024, ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv.BaseContext = func(net.Listener) context.Context { return ctx }
	result := make(chan error, 1)
	go func() { logger.Info("server listening", "addr", c.Addr); result <- srv.ListenAndServe() }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdown, cancel := context.WithTimeout(context.Background(), c.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			_ = srv.Close()
			return err
		}
		err := <-result
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
func setupToken(path string) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
			return "", fmt.Errorf("%s must be a regular file with permissions 0600", path)
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		existing := strings.TrimSpace(string(bytes))
		if len(existing) < 32 {
			return "", fmt.Errorf("existing setup token is too short")
		}
		return existing, nil
	}
	if err != nil {
		return "", err
	}
	_, writeErr := file.WriteString(token + "\n")
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return token, nil
}
