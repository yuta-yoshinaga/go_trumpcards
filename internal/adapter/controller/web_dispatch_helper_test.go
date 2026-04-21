//go:build test

package controller

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeResetStepLogger records which interactor method was invoked — the
// interactor surface is uniform across simple games, so a stub is enough to
// prove dispatchResetStepLog routes each command to the right method.
type fakeResetStepLogger struct {
	resetCalls int
	stepCalls  int
	logCalls   int
}

func (f *fakeResetStepLogger) Reset() string {
	f.resetCalls++
	return `{"method":"reset"}`
}

func (f *fakeResetStepLogger) Step() string {
	f.stepCalls++
	return `{"method":"step"}`
}

func (f *fakeResetStepLogger) ActionLog() string {
	f.logCalls++
	return `{"method":"log"}`
}

func TestDispatchResetStepLog(t *testing.T) {
	cases := []struct {
		cmd      string
		wantBody string
		wantFn   func(*fakeResetStepLogger) int
	}{
		{"r", `{"method":"reset"}`, func(f *fakeResetStepLogger) int { return f.resetCalls }},
		{"reset", `{"method":"reset"}`, func(f *fakeResetStepLogger) int { return f.resetCalls }},
		{"s", `{"method":"step"}`, func(f *fakeResetStepLogger) int { return f.stepCalls }},
		{"step", `{"method":"step"}`, func(f *fakeResetStepLogger) int { return f.stepCalls }},
		{"l", `{"method":"log"}`, func(f *fakeResetStepLogger) int { return f.logCalls }},
		{"log", `{"method":"log"}`, func(f *fakeResetStepLogger) int { return f.logCalls }},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			bc := &baseController{}
			w := httptest.NewRecorder()
			f := &fakeResetStepLogger{}
			ok := dispatchResetStepLog(tc.cmd, bc, w, f)
			assert.True(t, ok)
			assert.Equal(t, 1, tc.wantFn(f))
			assert.Equal(t, tc.wantBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestDispatchResetStepLog_Unknown(t *testing.T) {
	bc := &baseController{}
	w := httptest.NewRecorder()
	f := &fakeResetStepLogger{}
	ok := dispatchResetStepLog("unknown", bc, w, f)
	assert.False(t, ok)
	assert.Equal(t, 0, f.resetCalls+f.stepCalls+f.logCalls)
	assert.Empty(t, w.Body.String())
}

// counter returns a func() string that records invocations into *n and
// returns a fixed JSON body. Keeps the dispatchResetAndLog /
// dispatchResetHintAndLog tests compact.
func counter(n *int, body string) func() string {
	return func() string {
		*n++
		return body
	}
}

func TestDispatchResetAndLog(t *testing.T) {
	cases := []struct {
		cmd         string
		wantBody    string
		wantReset   int
		wantLog     int
		wantHandled bool
	}{
		{"r", `{"m":"reset"}`, 1, 0, true},
		{"reset", `{"m":"reset"}`, 1, 0, true},
		{"log", `{"m":"log"}`, 0, 1, true},
		{"l", `{"m":"log"}`, 0, 1, true},
		{"nope", "", 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			bc := &baseController{}
			w := httptest.NewRecorder()
			var rCalls, lCalls int
			ok := dispatchResetAndLog(tc.cmd, bc, w,
				counter(&rCalls, `{"m":"reset"}`),
				counter(&lCalls, `{"m":"log"}`))
			assert.Equal(t, tc.wantHandled, ok)
			assert.Equal(t, tc.wantReset, rCalls)
			assert.Equal(t, tc.wantLog, lCalls)
			assert.Equal(t, tc.wantBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestDispatchResetHintAndLog(t *testing.T) {
	cases := []struct {
		cmd                 string
		wantBody            string
		wantR, wantH, wantL int
		wantHandled         bool
	}{
		{"r", `{"m":"reset"}`, 1, 0, 0, true},
		{"reset", `{"m":"reset"}`, 1, 0, 0, true},
		{"h", `{"m":"hint"}`, 0, 1, 0, true},
		{"hint", `{"m":"hint"}`, 0, 1, 0, true},
		{"log", `{"m":"log"}`, 0, 0, 1, true},
		{"l", `{"m":"log"}`, 0, 0, 1, true},
		{"nope", "", 0, 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			bc := &baseController{}
			w := httptest.NewRecorder()
			var rCalls, hCalls, lCalls int
			ok := dispatchResetHintAndLog(tc.cmd, bc, w,
				counter(&rCalls, `{"m":"reset"}`),
				counter(&hCalls, `{"m":"hint"}`),
				counter(&lCalls, `{"m":"log"}`))
			assert.Equal(t, tc.wantHandled, ok)
			assert.Equal(t, tc.wantR, rCalls)
			assert.Equal(t, tc.wantH, hCalls)
			assert.Equal(t, tc.wantL, lCalls)
			assert.Equal(t, tc.wantBody, strings.TrimSpace(w.Body.String()))
		})
	}
}
