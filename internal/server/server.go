package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jeeftor/audiobook-organizer/internal/app"
)

//go:embed all:static
var embeddedStatic embed.FS

// Config contains HTTP server settings.
type Config struct {
	Token string
}

// Server provides local web UI HTTP routes.
type Server struct {
	config Config
	app    *app.Service
	static fs.FS
}

// New creates a server with embedded static assets.
func New(config Config, service *app.Service) (*Server, error) {
	static, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		return nil, err
	}
	return &Server{
		config: config,
		app:    service,
		static: static,
	}, nil
}

// Serve runs the server on the provided listener until shutdown.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	httpServer := &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err := httpServer.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// URL returns a display URL for the listener address.
func URL(host string, listener net.Listener, token string) string {
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return fmt.Sprintf("http://%s/?token=%s", listener.Addr().String(), token)
	}
	displayHost := host
	if displayHost == "" || displayHost == "0.0.0.0" || displayHost == "::" {
		displayHost = "127.0.0.1"
	}
	if strings.Contains(displayHost, ":") && !strings.HasPrefix(displayHost, "[") {
		displayHost = "[" + displayHost + "]"
	}
	return fmt.Sprintf("http://%s:%s/?token=%s", displayHost, port, token)
}
