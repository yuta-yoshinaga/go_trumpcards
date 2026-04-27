package web

import (
	"bytes"
	"testing"
)

func TestGetListenAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{
			name: "defaults to loopback and 8080 when neither env var is set",
			host: "",
			port: "",
			want: "127.0.0.1:8080",
		},
		{
			name: "honors HOST env var",
			host: "0.0.0.0",
			port: "",
			want: "0.0.0.0:8080",
		},
		{
			name: "honors PORT env var",
			host: "",
			port: "3000",
			want: "127.0.0.1:3000",
		},
		{
			name: "honors both HOST and PORT",
			host: "192.168.1.1",
			port: "9000",
			want: "192.168.1.1:9000",
		},
		{
			name: "HOST=0.0.0.0 exposes on all interfaces",
			host: "0.0.0.0",
			port: "8080",
			want: "0.0.0.0:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOST", tt.host)
			t.Setenv("PORT", tt.port)
			if got := getListenAddr(); got != tt.want {
				t.Errorf("getListenAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewTrumpCardsWeb_DefaultsForQuietMode guards issue #1452: quiet must
// default to false (interactive behavior preserved) and the stderr sink must
// be non-nil so Exec can write without NPE.
func TestNewTrumpCardsWeb_DefaultsForQuietMode(t *testing.T) {
	w := NewTrumpCardsWeb()
	if w.quiet {
		t.Error("quiet must default to false so interactive users still see the banner")
	}
	if w.stderr == nil {
		t.Error("stderr sink must be initialized; nil would NPE in Exec")
	}
}

// TestSetQuiet_TogglesFlag verifies the setter round-trips correctly. The
// actual gating of free-text output is exercised via the interactive vs
// non-interactive paths in the main package tests — here we only check the
// state of the flag so the unit test stays hermetic (no network / signals).
func TestSetQuiet_TogglesFlag(t *testing.T) {
	w := NewTrumpCardsWeb()
	w.SetQuiet(true)
	if !w.quiet {
		t.Fatal("SetQuiet(true) did not enable quiet mode")
	}
	w.SetQuiet(false)
	if w.quiet {
		t.Fatal("SetQuiet(false) did not disable quiet mode")
	}
}

// TestSetStderr_RedirectsSink verifies the free-text sink is injectable.
// Tests that don't want to read from os.Stderr (most of them) can point
// SetStderr at a bytes.Buffer and still cover the emit paths.
func TestSetStderr_RedirectsSink(t *testing.T) {
	w := NewTrumpCardsWeb()
	var buf bytes.Buffer
	w.SetStderr(&buf)
	if w.stderr != &buf {
		t.Fatal("SetStderr did not redirect the sink")
	}
}

// TestIsExposedHost pins down the predicate that decides whether a host
// string is "reachable from other machines" (i.e. non-loopback). Used by
// the network-exposure warning. See issue #1536.
func TestIsExposedHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		// Non-exposed (loopback / safe defaults).
		{"", false},          // empty → defaults applied elsewhere; treat as safe
		{"localhost", false}, // hostname; stays cautious without DNS resolution
		{"127.0.0.1", false},
		{"127.0.0.5", false}, // 127.0.0.0/8 is all loopback
		{"::1", false},
		// Exposed (IPv4/IPv6 unspecified or routable).
		{"0.0.0.0", true},      // IPv4 unspecified — listens on every interface
		{"::", true},           // IPv6 unspecified
		{"192.168.1.10", true}, // private LAN address
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"8.8.8.8", true}, // public — would not be a sensible bind, but classify as exposed
		// Unparseable — be cautious and treat as not exposed; we don't resolve DNS.
		{"myserver.local", false},
		{"not.a.real.host", false},
		{"192.168.x.y", false}, // bad IP literal
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isExposedHost(tt.host); got != tt.want {
				t.Errorf("isExposedHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

// TestMaybeWarnExposed_EmitsForNonLoopback verifies the warning fires for
// non-loopback hosts and stays silent for loopback / quiet mode. Asserts on
// the boundAddr being present in the message so users can grep for the
// specific listener that triggered the warning.
func TestMaybeWarnExposed_EmitsForNonLoopback(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		quiet    bool
		wantWarn bool
	}{
		{"0.0.0.0 warns", "0.0.0.0", false, true},
		{"127.0.0.1 silent", "127.0.0.1", false, false},
		{"::1 silent", "::1", false, false},
		{"localhost silent", "localhost", false, false},
		{"::  warns", "::", false, true},
		{"private LAN warns", "192.168.1.10", false, true},
		{"quiet suppresses 0.0.0.0", "0.0.0.0", true, false},
		{"quiet suppresses 192.168.x", "192.168.1.10", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewTrumpCardsWeb()
			var buf bytes.Buffer
			w.SetStderr(&buf)
			w.SetQuiet(tt.quiet)
			boundAddr := tt.host + ":8080"
			w.maybeWarnExposed(tt.host, boundAddr)
			gotWarn := buf.Len() > 0
			if gotWarn != tt.wantWarn {
				t.Errorf("warning emitted = %v, want %v (output=%q)", gotWarn, tt.wantWarn, buf.String())
			}
			if tt.wantWarn && !bytes.Contains(buf.Bytes(), []byte(boundAddr)) {
				t.Errorf("warning must include the bound addr %q so users can grep; got: %q",
					boundAddr, buf.String())
			}
		})
	}
}
