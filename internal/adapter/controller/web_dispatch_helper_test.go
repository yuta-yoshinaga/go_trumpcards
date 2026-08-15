//go:build test

package controller

import (
	"net/http/httptest"
	"strconv"
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

// dispatchTrickPlay consolidates 8 byte-identical per-game dispatchers
// (aluetteDispatch, ganjifaDispatch, klaverjasDispatch, manilleDispatch,
// mariasDispatch, sedmaDispatch, spoilFiveDispatch, suecaDispatch). They
// differed only in name and in the concrete interactor type — see #5368.
//
// Taking function values rather than an interface is what makes that possible:
// each game's ResetWithConfig takes its own config type, so no single interface
// can cover them. That is the same reason dispatchResetHintAndLog above takes
// resetFn/hintFn/actionLogFn.
func TestDispatchTrickPlay(t *testing.T) {
	cases := []struct {
		name     string
		cmd      string
		cardIdx  *int
		wantBody string
		wantCode int
	}{
		{"reset", "r", nil, `{"method":"reset"}`, 200},
		{"reset long form", "reset", nil, `{"method":"reset"}`, 200},
		{"play", "p", intPtr(2), `{"method":"play:2"}`, 200},
		{"play long form", "play", intPtr(0), `{"method":"play:0"}`, 200},
		{"next trick", "n", nil, `{"method":"next"}`, 200},
		{"next trick long form", "next", nil, `{"method":"next"}`, 200},
		{"next round", "nr", nil, `{"method":"nextround"}`, 200},
		{"next round long form", "nextround", nil, `{"method":"nextround"}`, 200},
		{"hint falls through to the shared helper", "h", nil, `{"method":"hint"}`, 200},
		{"log falls through to the shared helper", "log", nil, `{"method":"log"}`, 200},
		// The missing-parameter path must 400 rather than dereference nil.
		{"play without an index", "p", nil, "cardIndex is required", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			bc := &baseController{}
			handled := dispatchTrickPlay(tc.cmd, bc, w, trickPlayFns{
				resetWithConfig: func() string { return `{"method":"reset"}` },
				play:            func(i int) string { return `{"method":"play:` + strconv.Itoa(i) + `"}` },
				nextTrick:       func() string { return `{"method":"next"}` },
				nextRound:       func() string { return `{"method":"nextround"}` },
				hint:            func() string { return `{"method":"hint"}` },
				actionLog:       func() string { return `{"method":"log"}` },
			}, tc.cardIdx, func(msg string) any { return map[string]string{"message": msg} })
			assert.True(t, handled, "every command in this table is handled")
			assert.Equal(t, tc.wantCode, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), tc.wantBody),
				"body %q should contain %q", w.Body.String(), tc.wantBody)
		})
	}
}

// An unknown command must NOT be swallowed: the caller falls back to its own
// default branch, and reporting true here would silently drop the request.
func TestDispatchTrickPlay_UnknownCommandIsNotHandled(t *testing.T) {
	w := httptest.NewRecorder()
	handled := dispatchTrickPlay("frobnicate", &baseController{}, w, trickPlayFns{
		resetWithConfig: func() string { return "" },
		play:            func(int) string { return "" },
		nextTrick:       func() string { return "" },
		nextRound:       func() string { return "" },
		hint:            func() string { return "" },
		actionLog:       func() string { return "" },
	}, nil, func(msg string) any { return msg })
	assert.False(t, handled)
	assert.Equal(t, 200, w.Code, "nothing should have been written")
	assert.Empty(t, w.Body.String())
}

// dispatchBidTrickPlay is dispatchTrickPlay plus a bid step, consolidating
// 6 more byte-identical dispatchers: fortyFivesDispatch, napDispatch,
// preferenceDispatch, soloWhistDispatch, twentyNineDispatch, viraDispatch.
// Kept separate from dispatchTrickPlay rather than making bid optional: a
// nil-able function field would let a game silently accept "b" and do nothing.
func TestDispatchBidTrickPlay(t *testing.T) {
	cases := []struct {
		name     string
		cmd      string
		bid      *int
		cardIdx  *int
		wantBody string
		wantCode int
	}{
		{"reset", "r", nil, nil, `{"method":"reset"}`, 200},
		{"bid", "b", intPtr(3), nil, `{"method":"bid:3"}`, 200},
		{"bid long form", "bid", intPtr(0), nil, `{"method":"bid:0"}`, 200},
		{"play", "p", nil, intPtr(5), `{"method":"play:5"}`, 200},
		{"next trick", "n", nil, nil, `{"method":"next"}`, 200},
		{"next round", "nr", nil, nil, `{"method":"nextround"}`, 200},
		{"hint", "h", nil, nil, `{"method":"hint"}`, 200},
		{"log", "log", nil, nil, `{"method":"log"}`, 200},
		{"bid without a value", "b", nil, nil, "bid is required", 400},
		{"play without an index", "p", nil, nil, "cardIndex is required", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handled := dispatchBidTrickPlay(tc.cmd, &baseController{}, w, bidTrickPlayFns{
				resetWithConfig: func() string { return `{"method":"reset"}` },
				bid:             func(b int) string { return `{"method":"bid:` + strconv.Itoa(b) + `"}` },
				play:            func(i int) string { return `{"method":"play:` + strconv.Itoa(i) + `"}` },
				nextTrick:       func() string { return `{"method":"next"}` },
				nextRound:       func() string { return `{"method":"nextround"}` },
				hint:            func() string { return `{"method":"hint"}` },
				actionLog:       func() string { return `{"method":"log"}` },
			}, tc.bid, tc.cardIdx, func(msg string) any { return map[string]string{"message": msg} })
			assert.True(t, handled)
			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBody)
		})
	}
}

func TestDispatchBidTrickPlay_UnknownCommandIsNotHandled(t *testing.T) {
	w := httptest.NewRecorder()
	handled := dispatchBidTrickPlay("frobnicate", &baseController{}, w, bidTrickPlayFns{
		resetWithConfig: func() string { return "" }, bid: func(int) string { return "" },
		play: func(int) string { return "" }, nextTrick: func() string { return "" },
		nextRound: func() string { return "" }, hint: func() string { return "" },
		actionLog: func() string { return "" },
	}, nil, nil, func(msg string) any { return msg })
	assert.False(t, handled)
	assert.Empty(t, w.Body.String())
}

// dispatchRummyMeld consolidates the 4 identical canasta-family dispatchers:
// burracoDispatch, canastaDispatch, handAndFootDispatch, sambaDispatch.
// Falls through to dispatchLog (not dispatchHintAndLog) — these games have no
// hint command, and wiring one that does not exist would answer a hint request
// with a nil call.
func TestDispatchRummyMeld(t *testing.T) {
	cases := []struct {
		name     string
		cmd      string
		cardIdx  *int
		wantBody string
		wantCode int
	}{
		{"reset", "r", nil, `{"method":"reset"}`, 200},
		{"draw stock", "ds", nil, `{"method":"drawstock"}`, 200},
		{"draw stock long form", "drawstock", nil, `{"method":"drawstock"}`, 200},
		{"draw discard", "dd", nil, `{"method":"drawdiscard"}`, 200},
		{"meld", "m", nil, `{"method":"meld"}`, 200},
		{"skip meld", "sm", nil, `{"method":"skipmeld"}`, 200},
		{"discard", "d", intPtr(4), `{"method":"discard:4"}`, 200},
		{"go out", "go", nil, `{"method":"goout"}`, 200},
		{"next round", "nr", nil, `{"method":"nextround"}`, 200},
		{"log", "log", nil, `{"method":"log"}`, 200},
		{"discard without an index", "d", nil, "cardIndex is required", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handled := dispatchRummyMeld(tc.cmd, &baseController{}, w, rummyMeldFns{
				resetWithConfig: func() string { return `{"method":"reset"}` },
				drawFromStock:   func() string { return `{"method":"drawstock"}` },
				drawFromDiscard: func() string { return `{"method":"drawdiscard"}` },
				meld:            func() string { return `{"method":"meld"}` },
				skipMeld:        func() string { return `{"method":"skipmeld"}` },
				discard:         func(i int) string { return `{"method":"discard:` + strconv.Itoa(i) + `"}` },
				goOut:           func() string { return `{"method":"goout"}` },
				nextRound:       func() string { return `{"method":"nextround"}` },
				actionLog:       func() string { return `{"method":"log"}` },
			}, tc.cardIdx, func(msg string) any { return map[string]string{"message": msg} })
			assert.True(t, handled)
			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBody)
		})
	}
}

// "h"/"hint" must fall through unhandled: these games have no hint, and the
// caller's default branch is what answers.
func TestDispatchRummyMeld_HintIsNotHandled(t *testing.T) {
	w := httptest.NewRecorder()
	handled := dispatchRummyMeld("h", &baseController{}, w, rummyMeldFns{
		resetWithConfig: func() string { return "" }, drawFromStock: func() string { return "" },
		drawFromDiscard: func() string { return "" }, meld: func() string { return "" },
		skipMeld: func() string { return "" }, discard: func(int) string { return "" },
		goOut: func() string { return "" }, nextRound: func() string { return "" },
		actionLog: func() string { return "" },
	}, nil, func(msg string) any { return msg })
	assert.False(t, handled)
	assert.Empty(t, w.Body.String())
}

// dispatchTableauOnlyMove consolidates scorpionMoveDispatch,
// spideretteMoveDispatch and waspMoveDispatch — solitaires whose only legal
// move is tableau-to-tableau, so anything else is a 400 rather than a branch.
//
// Zones are passed as plain values, not as a struct: every game declares its
// own zone type (ScorpionWebZone, SpideretteWebZone, …) with identical fields,
// and there is no shared type to accept. The caller unpacks; the helper decides.
func TestDispatchTableauOnlyMove(t *testing.T) {
	col, idx := 1, 2

	cases := []struct {
		name                       string
		haveFrom, haveTo           bool
		fromZone, toZone           string
		fromCol, fromIdx, toColPtr *int
		wantBody                   string
		wantCode                   int
	}{
		{"tableau to tableau", true, true, "tableau", "tableau", &col, &idx, &col, `{"moved":true}`, 200},
		{"missing from", false, true, "", "tableau", nil, nil, &col, "from and to are required", 400},
		{"missing to", true, false, "tableau", "", &col, &idx, nil, "from and to are required", 400},
		{"foundation is not a legal target", true, true, "tableau", "foundation", &col, &idx, &col, "Only tableau to tableau", 400},
		{"waste is not a legal source", true, true, "waste", "tableau", &col, &idx, &col, "Only tableau to tableau", 400},
		{"missing cardIndex", true, true, "tableau", "tableau", &col, nil, &col, "from.cardIndex", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handled := dispatchTableauOnlyMove(&baseController{}, w, tableauMove{
				haveFrom: tc.haveFrom, haveTo: tc.haveTo,
				fromZone: tc.fromZone, toZone: tc.toZone,
				fromCol: tc.fromCol, fromCardIndex: tc.fromIdx, toCol: tc.toColPtr,
			}, func(fromCol, cardIndex, toCol int) string { return `{"moved":true}` },
				func(msg string) any { return map[string]string{"message": msg} })
			assert.True(t, handled, "the helper always answers -- it never falls through")
			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBody)
		})
	}
}

// The move function must receive the indices in the documented order; swapping
// fromCol and toCol renders a legal-looking move that plays the wrong card.
func TestDispatchTableauOnlyMove_PassesIndicesInOrder(t *testing.T) {
	from, ci, to := 3, 4, 5
	var got []int
	w := httptest.NewRecorder()
	dispatchTableauOnlyMove(&baseController{}, w, tableauMove{
		haveFrom: true, haveTo: true, fromZone: "tableau", toZone: "tableau",
		fromCol: &from, fromCardIndex: &ci, toCol: &to,
	}, func(fromCol, cardIndex, toCol int) string {
		got = []int{fromCol, cardIndex, toCol}
		return "{}"
	}, func(msg string) any { return msg })
	assert.Equal(t, []int{3, 4, 5}, got)
}
