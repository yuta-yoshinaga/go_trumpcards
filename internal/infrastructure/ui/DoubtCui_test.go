//go:build test

package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRunDoubtCountdown_TTYReturnsTrueOnDoubtInput exercises the happy path:
// the user types "d" while the countdown is running, so the helper must
// return true without consuming any timer ticks and must close the in-place
// prompt line with a newline so the next CUI line starts cleanly.
func TestRunDoubtCountdown_TTYReturnsTrueOnDoubtInput(t *testing.T) {
	var buf bytes.Buffer
	in := make(chan string, 1)
	ticks := make(chan struct{})

	in <- "d"
	got := runDoubtCountdown(&buf, 10, in, ticks, true)

	assert.True(t, got)
	out := buf.String()
	assert.Contains(t, out, "\a", "TTY mode must emit the bell on entry")
	assert.Contains(t, out, "[10s]", "initial prompt must show the full window")
	assert.True(t, strings.HasSuffix(out, "\n"), "must terminate the in-place line with a newline so subsequent output is not glued to the prompt")
}

// TestRunDoubtCountdown_TTYRedrawsEachTickAndTimesOut walks the countdown
// from windowSec=3 down to timeout, verifying that each tick rewrites the
// seconds in-place via \r and that the final tick prints the timeout banner.
func TestRunDoubtCountdown_TTYRedrawsEachTickAndTimesOut(t *testing.T) {
	var buf bytes.Buffer
	in := make(chan string)
	ticks := make(chan struct{}, 3)

	// 3 ticks consume the entire window (3 → 2 → 1 → timeout).
	ticks <- struct{}{}
	ticks <- struct{}{}
	ticks <- struct{}{}

	got := runDoubtCountdown(&buf, 3, in, ticks, true)

	assert.False(t, got)
	out := buf.String()
	assert.Contains(t, out, "[3s]", "must render the initial countdown value")
	assert.Contains(t, out, "[2s]", "must redraw after the first tick")
	assert.Contains(t, out, "[1s]", "must redraw on every tick down to 1")
	assert.NotContains(t, out, "[0s]", "must not render a zero-second prompt; the final tick is the timeout")
	assert.Contains(t, out, "\r", "TTY redraws must use carriage return")
	assert.Contains(t, out, "Timeout: skipping doubt", "timeout banner must appear after the final tick")
}

// TestRunDoubtCountdown_NonTTYPrintsPromptOnceAndNoBell verifies the
// fallback path: when stdout is not a terminal (CI logs, piped output)
// the helper prints the prompt once with a trailing newline, never emits
// \a or \r, and still drives the input/timeout state machine.
func TestRunDoubtCountdown_NonTTYPrintsPromptOnceAndNoBell(t *testing.T) {
	var buf bytes.Buffer
	in := make(chan string)
	ticks := make(chan struct{}, 1)
	ticks <- struct{}{} // 1 tick is the entire window

	got := runDoubtCountdown(&buf, 1, in, ticks, false)

	assert.False(t, got)
	out := buf.String()
	assert.NotContains(t, out, "\a", "non-TTY mode must not emit the terminal bell")
	assert.NotContains(t, out, "\r", "non-TTY mode must not use carriage returns")
	assert.Equal(t, 1, strings.Count(out, "[1s]"), "non-TTY must print the prompt exactly once")
	assert.Contains(t, out, "Timeout: skipping doubt")
}

// TestRunDoubtCountdown_DoubtKeywordAccepted treats "doubt" the same as "d",
// matching the existing handleDoubtWindow contract (case-sensitive, first
// whitespace-separated token).
func TestRunDoubtCountdown_DoubtKeywordAccepted(t *testing.T) {
	var buf bytes.Buffer
	in := make(chan string, 1)
	ticks := make(chan struct{})

	in <- "doubt now"
	got := runDoubtCountdown(&buf, 5, in, ticks, false)

	assert.True(t, got)
}

// TestRunDoubtCountdown_BlankInputDoesNotDoubt verifies that an empty
// Enter press resolves the window as "skip", matching the legacy behavior
// where only the literal "d"/"doubt" first token raises a doubt.
func TestRunDoubtCountdown_BlankInputDoesNotDoubt(t *testing.T) {
	var buf bytes.Buffer
	in := make(chan string, 1)
	ticks := make(chan struct{})

	in <- ""
	got := runDoubtCountdown(&buf, 5, in, ticks, false)

	assert.False(t, got)
}
