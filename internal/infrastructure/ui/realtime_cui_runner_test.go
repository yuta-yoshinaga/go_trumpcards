//go:build test

package ui

import (
	"bytes"
	"strings"
	"testing"

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
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
	assert.Contains(t, buf.String(), "fresh game")
	assert.Equal(t, []string{"r"}, me.calls)
}

func TestRealtimeCuiCore_TickInvokesTickCommand(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)
	ticks := make(chan struct{}, 2)
	ticks <- struct{}{}
	ticks <- struct{}{}
	close(ticks)
	close(keys)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "tick": "tock"}}
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
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
	close(ticks)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "j": "you slap", "s": "you step"}}
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
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
	close(ticks)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init"}}
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
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
	close(ticks)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init", "j": "slap"}}
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
	assert.Equal(t, []string{"r", "j"}, me.calls)
}

func TestRealtimeCuiCore_EOFEndsLoop(t *testing.T) {
	t.Parallel()
	keys := make(chan rune)
	close(keys) // EOF immediately
	ticks := make(chan struct{})
	close(ticks)
	var buf bytes.Buffer
	me := &realtimeMockExecer{response: map[string]string{"r": "init"}}
	realtimeCuiCore(me, keys, ticks, &buf, SlapjackRealtimeKeyMap)
	assert.Equal(t, []string{"r"}, me.calls)
}
