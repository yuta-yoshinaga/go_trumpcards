package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// gameEntry pairs a route name with its controller instance.
type gameEntry struct {
	name       string
	controller games.WebController
}

// TrumpCardsWeb トランプカードゲームWebクラス
type TrumpCardsWeb struct {
	games  []gameEntry
	quiet  bool      // when true, suppress human-friendly startup/shutdown messages
	stderr io.Writer // injectable for tests; defaults to os.Stderr
}

// NewTrumpCardsWeb コンストラクタ
func NewTrumpCardsWeb() *TrumpCardsWeb {
	web := &TrumpCardsWeb{stderr: os.Stderr}
	web.registerAll()
	return web
}

// SetQuiet toggles suppression of the free-text startup/shutdown messages.
// Structured slog logs are emitted regardless — this setter only controls
// the human-friendly lines meant for interactive terminals. See issue #1452.
func (web *TrumpCardsWeb) SetQuiet(v bool) {
	web.quiet = v
}

// SetStderr overrides the stderr sink used for free-text messages. Test
// helper; production code relies on the default os.Stderr.
func (web *TrumpCardsWeb) SetStderr(w io.Writer) {
	web.stderr = w
}

// registerAll builds the per-game controllers from the central registry in
// internal/infrastructure/games. Adding a game there automatically exposes
// it on /<name>/exec here — there is no separate list to maintain.
func (web *TrumpCardsWeb) registerAll() {
	for _, g := range games.All() {
		web.games = append(web.games, gameEntry{
			name:       g.Name,
			controller: g.NewWebController(),
		})
	}
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() error {
	mux := http.NewServeMux()

	for _, g := range web.games {
		mux.HandleFunc("POST /"+g.name+"/exec", g.controller.Exec)
	}
	RegisterSwaggerRoutes(mux)
	mux.Handle("/", http.FileServer(http.Dir("public")))

	// Apply CORS middleware if allowed origins are configured.
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOriginsStr == "" && os.Getenv("APP_ENV") != "production" {
		allowedOriginsStr = "http://localhost:5173,http://localhost:8080"
	}
	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(allowedOriginsStr); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	const (
		readTimeout     = 10 * time.Second
		writeTimeout    = 30 * time.Second
		idleTimeout     = 60 * time.Second
		shutdownTimeout = 30 * time.Second
	)
	srv := &http.Server{
		Addr:         getListenAddr(),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		slog.Error("server listen error", "error", err)
		return fmt.Errorf("failed to listen on %s: %w", srv.Addr, err)
	}
	// Free-text messages go to the configurable stderr sink (defaults to
	// os.Stderr). Structured slog events fire regardless so systemd / docker
	// / log shippers still see every lifecycle event. See issue #1452.
	boundAddr := ln.Addr().String()
	slog.Info("web server listening", "addr", boundAddr)
	if !web.quiet {
		_, _ = fmt.Fprintln(web.stderr, i18n.Tf("webServerRunning", "addr", boundAddr))
		_, _ = fmt.Fprintln(web.stderr, i18n.T("webServerStop"))
	}
	// Network-exposure warning fires only after the listening banner so the
	// user sees the bound address first, then "by the way, this is reachable
	// from your LAN". Suppressed by --quiet / TRUMPCARDS_QUIET=1 / non-TTY
	// stderr (the latter via the caller's quiet override). See issue #1536.
	if host, _, splitErr := net.SplitHostPort(boundAddr); splitErr == nil {
		web.maybeWarnExposed(host, boundAddr)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	var runErr error
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			runErr = fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		if !web.quiet {
			_, _ = fmt.Fprintln(web.stderr, "\n"+i18n.T("webServerShutdown"))
		}
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}

	for _, g := range web.games {
		g.controller.Stop()
	}
	if !web.quiet {
		_, _ = fmt.Fprintln(web.stderr, i18n.T("webServerStopped"))
	}
	slog.Info("server stopped")
	return runErr
}

// getListenAddr returns the "host:port" address the web server should bind to.
// Default is 127.0.0.1:8080; set HOST=0.0.0.0 to expose on all interfaces.
func getListenAddr() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return net.JoinHostPort(host, port)
}

// isExposedHost reports whether host is reachable from other machines on the
// network — i.e. anything that is not a loopback address. Used to drive the
// network-exposure warning when the user binds to 0.0.0.0 or a LAN IP. We
// deliberately do NOT resolve DNS hostnames: an unparseable host falls back
// to "not exposed" so a typo or non-routable name (myserver.local) does not
// trigger a misleading warning. See issue #1536.
func isExposedHost(host string) bool {
	switch host {
	case "", "localhost":
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	// IsUnspecified covers 0.0.0.0 and ::, which listen on every interface.
	// IsLoopback covers 127.0.0.0/8 and ::1.
	return ip.IsUnspecified() || !ip.IsLoopback()
}

// maybeWarnExposed emits a one-line warning to web.stderr when the bound
// address is reachable from outside the local machine. It is a no-op when
// quiet mode is set or the host is loopback. Structured slog logs are not
// affected — this is purely a free-text helper for interactive terminals.
// See issue #1536.
func (web *TrumpCardsWeb) maybeWarnExposed(host, boundAddr string) {
	if web.quiet || !isExposedHost(host) {
		return
	}
	_, _ = fmt.Fprintln(web.stderr, i18n.Tf("webServerExposureWarning", "addr", boundAddr))
}
