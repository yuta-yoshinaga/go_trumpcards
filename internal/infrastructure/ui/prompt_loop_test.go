//go:build test

package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// promptMockExecer implements CuiExecer for testing.
type promptMockExecer struct {
	calls   []string
	results []string
	idx     int
}

func (m *promptMockExecer) Exec(command string) string {
	m.calls = append(m.calls, command)
	if m.idx < len(m.results) {
		r := m.results[m.idx]
		m.idx++
		return r
	}
	return "unexpected call"
}

func TestHandlePromptLoop_NoPrompt(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader(""), &buf)
	me := &promptMockExecer{}
	result := handlePromptLoop(reader, me, "normal result", "", &buf)
	assert.Equal(t, "normal result", result)
	assert.Empty(t, me.calls)
}

func TestHandlePromptLoop_SinglePrompt(t *testing.T) {
	t.Parallel()
	// Simulate: controller returns prompt, user types "100", controller returns "bet ok"
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader("100\n"), &buf)
	me := &promptMockExecer{results: []string{"bet ok"}}
	prompt := cuiutil.PromptRequest("Enter bet amount:", "b {0}")
	result := handlePromptLoop(reader, me, prompt, "", &buf)
	assert.Equal(t, "bet ok", result)
	assert.Equal(t, []string{"b 100"}, me.calls)
	assert.Contains(t, buf.String(), "Enter bet amount:")
}

func TestHandlePromptLoop_ChainedPrompts(t *testing.T) {
	t.Parallel()
	// Simulate wizard: m -> prompt source zone -> user types "t" -> prompt column -> user types "3" -> "move ok"
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader("t\n3\n"), &buf)
	secondPrompt := cuiutil.PromptRequest("Enter column:", "m t {0}")
	me := &promptMockExecer{results: []string{secondPrompt, "move ok"}}
	firstPrompt := cuiutil.PromptRequest("Enter source zone:", "m {0}")
	result := handlePromptLoop(reader, me, firstPrompt, "klondike", &buf)
	assert.Equal(t, "move ok", result)
	assert.Equal(t, []string{"m t", "m t 3"}, me.calls)
}

func TestHandlePromptLoop_EmptyInput_Cancels(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader("\n"), &buf)
	me := &promptMockExecer{}
	prompt := cuiutil.PromptRequest("Enter amount:", "b {0}")
	result := handlePromptLoop(reader, me, prompt, "", &buf)
	assert.Equal(t, i18n.T("cancelled"), result)
	assert.Empty(t, me.calls)
}

func TestHandlePromptLoop_EOF_ReturnsQuit(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader(""), &buf)
	me := &promptMockExecer{}
	prompt := cuiutil.PromptRequest("Enter amount:", "b {0}")
	result := handlePromptLoop(reader, me, prompt, "", &buf)
	assert.Equal(t, i18n.QuitSentinel, result)
	assert.Empty(t, me.calls)
}

// TestHandlePromptLoop_GameNameInPrompt verifies issue #1605: the gameName
// passed to handlePromptLoop threads through to the per-prompt readInput so
// chained-prompt input lines also carry the "[gameName] > " context. This
// matches the top-level readInput contract (issue #1605 makes the
// single-game-mode loop use the same gameName as interactive mode).
func TestHandlePromptLoop_GameNameInPrompt(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader("100\n"), &buf)
	me := &promptMockExecer{results: []string{"bet ok"}}
	prompt := cuiutil.PromptRequest("Enter bet amount:", "b {0}")
	_ = handlePromptLoop(reader, me, prompt, "blackjack", &buf)
	assert.Contains(t, buf.String(), "[blackjack] > ", "prompt must include gameName context")
}

func TestHandlePromptLoop_MalformedPrompt(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	reader := newScannerLineReader(strings.NewReader(""), &buf)
	me := &promptMockExecer{}
	// Malformed: no tab separator, so template is empty
	malformed := "PROMPT:Just a message"
	result := handlePromptLoop(reader, me, malformed, "", &buf)
	assert.Equal(t, "Just a message", result)
	assert.Empty(t, me.calls)
}
