//go:build test

package infrastructure

import (
	"bytes"
	"log/slog"
	"testing"
)

// The default format follows the stream, not a constant.
//
// `trumpcards web` prints a human-readable banner to stderr and slog fires the
// same lifecycle event alongside it. With a JSON handler pinned regardless of
// destination, a terminal user reads the address twice — once as
// `{"time":…,"msg":"web server listening","addr":…}` and once as prose — and on
// a failed bind reads the same failure twice. Issue #5357.
//
// The non-TTY half is the constraint that must not regress: issue #1452 made
// the structured events fire unconditionally so systemd / docker / log shippers
// see every lifecycle event. Those consumers are never a terminal, so keying
// the default on the stream preserves them exactly.
func TestLogHandlerFor_DefaultFollowsTheStream(t *testing.T) {
	tests := []struct {
		name  string
		isTTY bool
		want  string
	}{
		{"terminal gets text", true, "*slog.TextHandler"},
		{"pipe or file gets JSON", false, "*slog.JSONHandler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := logHandlerFor(&bytes.Buffer{}, "", slog.LevelInfo, tt.isTTY)
			if got := handlerName(h); got != tt.want {
				t.Errorf("logHandlerFor(format=\"\", isTTY=%v) = %s, want %s", tt.isTTY, got, tt.want)
			}
		})
	}
}

// An explicit LOG_FORMAT is a deliberate choice and outranks the stream, in
// both directions — including forcing JSON onto a terminal, which is how you
// reproduce a shipper's view while debugging interactively.
func TestLogHandlerFor_ExplicitFormatWinsOverTheStream(t *testing.T) {
	tests := []struct {
		name   string
		format string
		isTTY  bool
		want   string
	}{
		{"text forced on a pipe", "text", false, "*slog.TextHandler"},
		{"json forced on a terminal", "json", true, "*slog.JSONHandler"},
		{"case is ignored", "TEXT", false, "*slog.TextHandler"},
		{"unrecognised value falls back to the stream", "yaml", true, "*slog.TextHandler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := logHandlerFor(&bytes.Buffer{}, tt.format, slog.LevelInfo, tt.isTTY)
			if got := handlerName(h); got != tt.want {
				t.Errorf("logHandlerFor(format=%q, isTTY=%v) = %s, want %s", tt.format, tt.isTTY, got, tt.want)
			}
		})
	}
}

// The level is orthogonal to the format: switching a terminal to text must not
// quietly start dropping records.
func TestLogHandlerFor_LevelIsIndependentOfFormat(t *testing.T) {
	for _, isTTY := range []bool{true, false} {
		var buf bytes.Buffer
		h := logHandlerFor(&buf, "", slog.LevelWarn, isTTY)
		if h.Enabled(t.Context(), slog.LevelInfo) {
			t.Errorf("isTTY=%v: INFO enabled at WARN level", isTTY)
		}
		if !h.Enabled(t.Context(), slog.LevelError) {
			t.Errorf("isTTY=%v: ERROR disabled at WARN level", isTTY)
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"nonsense", slog.LevelInfo},
	}
	for _, tt := range tests {
		if got := parseLogLevel(tt.in); got != tt.want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// handlerName reports the concrete handler type, which is what distinguishes
// the two formats. Comparing rendered output instead would couple the test to
// slog's exact framing.
func handlerName(h slog.Handler) string {
	switch h.(type) {
	case *slog.TextHandler:
		return "*slog.TextHandler"
	case *slog.JSONHandler:
		return "*slog.JSONHandler"
	default:
		return "unknown"
	}
}

// The default level follows the stream for the same reason the format does:
// on a terminal the INFO lifecycle events restate the human-readable banner.
func TestLogLevelFor(t *testing.T) {
	tests := []struct {
		name  string
		env   string
		isTTY bool
		want  slog.Level
	}{
		{"terminal defaults to WARN so the banner is not echoed", "", true, slog.LevelWarn},
		{"pipe keeps INFO for log shippers", "", false, slog.LevelInfo},
		{"explicit level wins on a terminal", "DEBUG", true, slog.LevelDebug},
		{"explicit level wins on a pipe", "ERROR", false, slog.LevelError},
		{"explicit INFO restores lifecycle events on a terminal", "INFO", true, slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logLevelFor(tt.env, tt.isTTY); got != tt.want {
				t.Errorf("logLevelFor(%q, isTTY=%v) = %v, want %v", tt.env, tt.isTTY, got, tt.want)
			}
		})
	}
}
