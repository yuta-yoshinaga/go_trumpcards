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
