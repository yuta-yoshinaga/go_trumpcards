// Package web — browser opener helper for `trumpcards web --open` (issue #1607).
//
// The opener is intentionally best-effort: a failure to spawn xdg-open / open
// / cmd /c start does not affect the running server. The caller decides how to
// surface the error (or whether to skip opening at all in non-TTY contexts).
package web

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// BrowserURLFor returns the URL `--open` should hand to the platform browser
// opener for a listener bound at boundAddr ("host:port" from net.Listen). The
// host is rewritten to 127.0.0.1 when it would not be reachable from the
// browser as-is — namely the unspecified addresses 0.0.0.0 / ::, which a
// server can listen on but a client cannot meaningfully connect to. The port
// is preserved verbatim so `--port 0`'s OS-assigned ephemeral port flows
// through unchanged. Returns "" if boundAddr cannot be parsed.
func BrowserURLFor(boundAddr string) string {
	host, port, err := net.SplitHostPort(boundAddr)
	if err != nil {
		return ""
	}
	switch host {
	case "0.0.0.0", "::", "":
		host = "127.0.0.1"
	}
	if strings.Contains(host, ":") {
		// IPv6 literal — bracket it so net/url-style consumers parse correctly.
		host = "[" + host + "]"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

// OpenBrowser launches the OS-default browser opener for url. It returns the
// command's Start() error so the caller can decide whether to log it; the
// process is detached (Start, not Run) so a long-lived browser does not block
// the caller. Best-effort by design — see package doc.
func OpenBrowser(url string) error {
	cmd, args := openerCommand(url)
	if cmd == "" {
		return fmt.Errorf("no known browser opener for GOOS=%q", runtime.GOOS)
	}
	return exec.Command(cmd, args...).Start()
}

// openerCommand returns the platform-appropriate opener and arguments. Split
// out from OpenBrowser so tests can verify the wiring without spawning a
// process.
func openerCommand(url string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{url}
	case "windows":
		// `cmd /c start "" <url>` — the empty title argument prevents `start`
		// from interpreting a URL with spaces as the window title.
		return "cmd", []string{"/c", "start", "", url}
	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly":
		return "xdg-open", []string{url}
	}
	return "", nil
}
