//go:build test

package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
