package web

import (
	"runtime"
	"testing"
)

// TestBrowserURLFor verifies issue #1607: the URL we hand to the OS opener
// rewrites unspecified addresses to loopback (a server can listen on 0.0.0.0
// but a browser cannot connect to it as-is) while preserving real IPv4 / IPv6
// hosts and the bound port (which may be ephemeral when --port 0).
func TestBrowserURLFor(t *testing.T) {
	tests := []struct {
		name      string
		boundAddr string
		want      string
	}{
		{"loopback ipv4 preserved", "127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"unspecified ipv4 -> loopback", "0.0.0.0:3000", "http://127.0.0.1:3000"},
		{"unspecified ipv6 -> loopback", "[::]:3000", "http://127.0.0.1:3000"},
		{"loopback ipv6 bracketed", "[::1]:9000", "http://[::1]:9000"},
		{"lan ipv4 preserved", "192.168.1.10:8080", "http://192.168.1.10:8080"},
		{"ephemeral port preserved", "127.0.0.1:54321", "http://127.0.0.1:54321"},
		{"unparseable returns empty", "not-an-addr", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BrowserURLFor(tt.boundAddr)
			if got != tt.want {
				t.Errorf("BrowserURLFor(%q) = %q, want %q", tt.boundAddr, got, tt.want)
			}
		})
	}
}

// TestOpenerCommand checks that the platform dispatch picks the conventional
// opener for the current GOOS without spawning a process. We intentionally
// only assert on the host platform — exhaustively faking GOOS would force a
// global-state knob (runtime.GOOS is read-only) for very little payoff.
func TestOpenerCommand(t *testing.T) {
	cmd, args := openerCommand("http://127.0.0.1:8080")
	switch runtime.GOOS {
	case "darwin":
		if cmd != "open" || len(args) != 1 || args[0] != "http://127.0.0.1:8080" {
			t.Errorf("darwin opener: got (%q, %v); want (\"open\", [<url>])", cmd, args)
		}
	case "windows":
		// `cmd /c start "" <url>` — the empty title arg is required.
		want := []string{"/c", "start", "", "http://127.0.0.1:8080"}
		if cmd != "cmd" || !equalSlice(args, want) {
			t.Errorf("windows opener: got (%q, %v); want (\"cmd\", %v)", cmd, args, want)
		}
	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly":
		if cmd != "xdg-open" || len(args) != 1 || args[0] != "http://127.0.0.1:8080" {
			t.Errorf("unix opener: got (%q, %v); want (\"xdg-open\", [<url>])", cmd, args)
		}
	default:
		// Unknown platform: opener should be empty so OpenBrowser surfaces a
		// friendly error rather than launching something unexpected.
		if cmd != "" {
			t.Errorf("unknown GOOS=%q should fall through to no opener; got %q", runtime.GOOS, cmd)
		}
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
