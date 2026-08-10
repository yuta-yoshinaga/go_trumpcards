//go:build test

package ui

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// erroringByteReader returns a fixed non-EOF error after streaming zero
// bytes. Used to exercise the readRealtimeKeys contract that non-EOF read
// failures are surfaced on errCh while keys is still closed cleanly.
type erroringByteReader struct{ err error }

func (e *erroringByteReader) Read(_ []byte) (int, error) { return 0, e.err }

// TestReadRealtimeKeys_EOFLeavesErrChEmpty verifies that a clean EOF stops
// the reader without writing to errCh — the caller must see no error and
// can return exit 0.
func TestReadRealtimeKeys_EOFLeavesErrChEmpty(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)
	errCh := make(chan error, 1)
	go readRealtimeKeys(&erroringByteReader{err: io.EOF}, keys, errCh)
	// Drain keys; the goroutine must close it without sending anything.
	for range keys {
		t.Fatalf("expected no key bytes on EOF reader")
	}
	select {
	case got := <-errCh:
		t.Fatalf("EOF must not surface on errCh, got %v", got)
	default:
	}
}

// TestReadRealtimeKeys_NonEOFErrorSurfacesOnErrCh verifies the contract the
// realtime runner relies on for exit code 1: a non-EOF read failure both
// closes keys (so realtimeCuiCore returns) and lands on errCh (so the
// runner can map it to exit 1) — addresses gemini's "MUST" feedback.
func TestReadRealtimeKeys_NonEOFErrorSurfacesOnErrCh(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("pty hung up")
	keys := make(chan rune)
	errCh := make(chan error, 1)
	go readRealtimeKeys(&erroringByteReader{err: sentinel}, keys, errCh)
	for range keys {
		t.Fatalf("expected no key bytes on erroring reader")
	}
	select {
	case got := <-errCh:
		assert.ErrorIs(t, got, sentinel)
	case <-time.After(time.Second):
		t.Fatal("non-EOF reader error did not surface on errCh")
	}
}

// realtimeMockExecer records every command Exec receives and returns
// canned responses. Unmatched calls return an empty string so the
// realtime loop's tick spam does not produce spurious failures.
type realtimeMockExecer struct {
	calls    []string
	response map[string]string
}

func (m *realtimeMockExecer) Exec(command string) string {
	m.calls = append(m.calls, command)
	if r, ok := m.response[command]; ok {
		return r
	}
	return ""
}

// neverQuit returns a channel that is never closed — used by tests that
// don't want the quit signal path to influence the loop.
func neverQuit() <-chan struct{} { return make(chan struct{}) }

// SlapjackRealtimeKeyMap is the canonical key→command mapping the realtime
// runner uses for Slapjack and Egyptian Ratscrew. Tests reach for it
// directly so the implementation cannot drift from the contract.
func TestSlapjackRealtimeKeyMap_HasExpectedKeys(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "j", SlapjackRealtimeKeyMap[' '], "space → slap")
	assert.Equal(t, "s", SlapjackRealtimeKeyMap['s'])
	assert.Equal(t, "s", SlapjackRealtimeKeyMap['S'])
	assert.Equal(t, "r", SlapjackRealtimeKeyMap['r'])
	assert.Equal(t, "q", SlapjackRealtimeKeyMap['q'])
	assert.Equal(t, "log", SlapjackRealtimeKeyMap['l'])
}

// The legend and the key map must describe each other exactly. `log`/`l` was
// mapped for months while the hand-written banner never mentioned it, and the
// banner named commands (`j`, `tick`, `sd <n>`) that no key produced. Checking
// only one direction would have missed one of those two failures, so both are
// asserted here. See issue #5179.
func TestRealtimeLegendCoversKeyMap(t *testing.T) {
	t.Parallel()

	// Every command reachable from the key map has a label...
	for key, cmd := range SlapjackRealtimeKeyMap {
		_, ok := realtimeCommandLabelKeys[cmd]
		assert.Truef(t, ok, "key %q maps to command %q which has no label key", string(key), cmd)
	}
	// ...and every label is reachable from the key map.
	reachable := make(map[string]bool, len(SlapjackRealtimeKeyMap))
	for _, cmd := range SlapjackRealtimeKeyMap {
		reachable[cmd] = true
	}
	for cmd := range realtimeCommandLabelKeys {
		assert.Truef(t, reachable[cmd], "command %q has a label but no key maps to it", cmd)
	}

	// And the rendered legend actually mentions each key.
	legend := strings.Join(realtimeLegendLines(SlapjackRealtimeKeyMap), "\n")
	for key := range SlapjackRealtimeKeyMap {
		assert.Containsf(t, legend, realtimeKeyLabel(key), "legend omits key %q", string(key))
	}
}

// Negative control for the round-trip above: a key added without a label must
// fail, otherwise the guard proves nothing.
func TestRealtimeLegendCoversKeyMap_CatchesUnlabelledKey(t *testing.T) {
	t.Parallel()
	rogue := map[rune]string{'z': "totally-new-command"}
	_, ok := realtimeCommandLabelKeys[rogue['z']]
	assert.False(t, ok, "fixture must be an unlabelled command for this control to mean anything")
	// realtimeLegendLines skips it rather than rendering a bogus line...
	legend := strings.Join(realtimeLegendLines(rogue), "\n")
	assert.NotContains(t, legend, "totally-new-command")
	// ...which is exactly what the round-trip assertion above would catch.
}

// The help key must never be dispatched to the controller: it is a loop
// concern, and "\x00help" is not a command any controller understands.
func TestRealtimeCuiCore_HelpKeyPrintsLegendWithoutDispatching(t *testing.T) {
	t.Parallel()
	keys := make(chan rune, 2)
	keys <- 'h'
	keys <- 'q'
	close(keys)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "fresh"}}
	realtimeCuiCore(me, keys, nil, neverQuit(), &buf, SlapjackRealtimeKeyMap)

	assert.Equal(t, []string{"r"}, me.calls, "help must not reach the controller")
	assert.Contains(t, buf.String(), i18n.T("realtime.labelSlap"))
	assert.Contains(t, buf.String(), i18n.T("realtime.labelLog"))
}

// The difficulty keys carry an argument, so they must arrive at the controller
// as a full "sd <n>" command string.
func TestRealtimeCuiCore_DifficultyKeysDispatchWithArgument(t *testing.T) {
	t.Parallel()
	keys := make(chan rune, 4)
	for _, k := range []rune{'1', '2', '3'} {
		keys <- k
	}
	close(keys)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "fresh"}}
	realtimeCuiCore(me, keys, nil, neverQuit(), &buf, SlapjackRealtimeKeyMap)

	assert.Equal(t, []string{"r", "sd 0", "sd 1", "sd 2"}, me.calls)
}

func TestRealtimeCuiCore_PrintsInitialResetOnStart(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)
	ticks := make(chan struct{})
	close(keys)
	close(ticks)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "fresh game"}}
	realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
	assert.Contains(t, buf.String(), "fresh game")
	assert.Equal(t, []string{"r"}, me.calls)
}

func TestRealtimeCuiCore_TickInvokesTickCommand(t *testing.T) {
	t.Parallel()
	// keys is intentionally left open until ticks have drained — closing
	// keys would terminate the loop before tick processing per the
	// production contract (stdin EOF == quit).
	keys := make(chan rune)
	ticks := make(chan struct{}, 2)
	ticks <- struct{}{}
	ticks <- struct{}{}
	// Close ticks after handing off the buffered values so the core sees
	// "no more ticks" once both have been consumed; then close keys to
	// drive loop termination.
	go func() {
		// give the core a moment to drain the buffered ticks before the
		// keys close races with them
		time.Sleep(50 * time.Millisecond)
		close(ticks)
		close(keys)
	}()
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "tick": "tock"}}
	realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
	assert.Equal(t, []string{"r", "tick", "tick"}, me.calls)
	assert.Equal(t, 2, strings.Count(buf.String(), "tock"))
}

func TestRealtimeCuiCore_KeyTriggersMappedCommand(t *testing.T) {
	t.Parallel()
	keys := make(chan rune, 2)
	keys <- ' '
	keys <- 's'
	close(keys)
	ticks := make(chan struct{})
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "j": "you slap", "s": "you step"}}
	realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
	assert.Equal(t, []string{"r", "j", "s"}, me.calls)
	out := buf.String()
	assert.Contains(t, out, "you slap")
	assert.Contains(t, out, "you step")
}

func TestRealtimeCuiCore_QKeyEndsLoop(t *testing.T) {
	t.Parallel()
	keys := make(chan rune, 3)
	keys <- ' '
	keys <- 'q'
	keys <- 's' // should never be processed
	close(keys)
	ticks := make(chan struct{})
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init"}}
	realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
	// 'q' must terminate before 's' is processed.
	assert.Equal(t, []string{"r", "j"}, me.calls)
}

func TestRealtimeCuiCore_UnknownKeyIsIgnored(t *testing.T) {
	t.Parallel()
	keys := make(chan rune, 2)
	keys <- 'z' // not in map
	keys <- ' '
	close(keys)
	ticks := make(chan struct{})
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "j": "slap"}}
	realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
	assert.Equal(t, []string{"r", "j"}, me.calls)
}

// TestRealtimeCuiCore_EOFEndsLoop verifies the production-realistic
// scenario: the ticker is still running when stdin closes. The loop
// must return promptly even though `ticks` is never closed.
func TestRealtimeCuiCore_EOFEndsLoop(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)
	close(keys)                  // EOF immediately
	ticks := make(chan struct{}) // intentionally left open (mirrors production)
	done := make(chan struct{})
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init"}}
	go func() {
		realtimeCuiCore(me, keys, ticks, neverQuit(), &buf, SlapjackRealtimeKeyMap)
		close(done)
	}()
	select {
	case <-done:
		// loop terminated as expected
	case <-time.After(time.Second):
		t.Fatal("realtimeCuiCore did not terminate on stdin EOF — would hang the runner")
	}
	assert.Equal(t, []string{"r"}, me.calls)
}

// TestRealtimeCuiCore_QuitChannelEndsLoop verifies signal-based exit:
// closing the quit channel must stop the loop the same way a 'q' key
// would. This is what RunRealtimeCuiLoop relies on for SIGINT/SIGTERM.
func TestRealtimeCuiCore_QuitChannelEndsLoop(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)      // open
	ticks := make(chan struct{}) // open
	quit := make(chan struct{})
	close(quit) // signal already fired
	done := make(chan struct{})
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init"}}
	go func() {
		realtimeCuiCore(me, keys, ticks, quit, &buf, SlapjackRealtimeKeyMap)
		close(done)
	}()
	select {
	case <-done:
		// loop terminated as expected
	case <-time.After(time.Second):
		t.Fatal("realtimeCuiCore did not terminate when quit channel closed")
	}
	assert.Equal(t, []string{"r"}, me.calls)
}
