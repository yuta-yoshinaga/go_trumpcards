//go:build test && (!js || !wasm || casino)

package controller_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// One distinct reply per interactor method: the assertions compare the reply, so
// a dispatcher that routed "f" to Play() (or "h" to ActionLog()) fails instead of
// returning the one shared payload every method would otherwise hand back.
const (
	tcrCuiResetReply = "reset result"
	tcrCuiBetReply   = "bet result"
	tcrCuiRebetReply = "rebet result"
	tcrCuiPlayReply  = "play result"
	tcrCuiFoldReply  = "fold result"
	tcrCuiHintReply  = "hint result"
	tcrCuiLogReply   = "action log result"
)

// newThreeCardRummyCuiFixture stubs every method. Bet's arguments are matched
// loosely; their values are asserted by
// TestThreeCardRummyCuiController_BetForwardsBothStakes.
func newThreeCardRummyCuiFixture(t *testing.T) (*usecase.MockThreeCardRummyInteractor, *controller.ThreeCardRummyCuiController) {
	t.Helper()
	m := new(usecase.MockThreeCardRummyInteractor)
	m.On("Reset").Return(tcrCuiResetReply)
	m.On("Bet", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(tcrCuiBetReply)
	m.On("Rebet").Return(tcrCuiRebetReply)
	m.On("Play").Return(tcrCuiPlayReply)
	m.On("Fold").Return(tcrCuiFoldReply)
	m.On("Hint").Return(tcrCuiHintReply)
	m.On("ActionLog").Return(tcrCuiLogReply)
	return m, controller.NewThreeCardRummyCuiController(m)
}

// newBareThreeCardRummyCuiController returns a controller over a mock with no
// expectations at all: reaching any interactor method blows up, so the
// "was not dispatched" assertions below cannot pass vacuously.
func newBareThreeCardRummyCuiController(t *testing.T) (*usecase.MockThreeCardRummyInteractor, *controller.ThreeCardRummyCuiController) {
	t.Helper()
	m := new(usecase.MockThreeCardRummyInteractor)
	return m, controller.NewThreeCardRummyCuiController(m)
}

// TestThreeCardRummyCuiController_Dispatch pins each typed command, long form
// and short alias alike, to the interactor method it must reach.
func TestThreeCardRummyCuiController_Dispatch(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		method string
	}{
		{"r", tcrCuiResetReply, "Reset"},
		{"reset", tcrCuiResetReply, "Reset"},
		{"b 100", tcrCuiBetReply, "Bet"},
		{"bet 100", tcrCuiBetReply, "Bet"},
		{"b 100 50", tcrCuiBetReply, "Bet"},
		{"bet 100 50", tcrCuiBetReply, "Bet"},
		{"rb", tcrCuiRebetReply, "Rebet"},
		{"rebet", tcrCuiRebetReply, "Rebet"},
		{"p", tcrCuiPlayReply, "Play"},
		{"play", tcrCuiPlayReply, "Play"},
		{"f", tcrCuiFoldReply, "Fold"},
		{"fold", tcrCuiFoldReply, "Fold"},
		{"h", tcrCuiHintReply, "Hint"},
		{"hint", tcrCuiHintReply, "Hint"},
		{"log", tcrCuiLogReply, "ActionLog"},
		{"l", tcrCuiLogReply, "ActionLog"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m, c := newThreeCardRummyCuiFixture(t)
			assert.Equal(t, tt.want, c.Exec(tt.input))
			m.AssertNumberOfCalls(t, tt.method, 1)
			assert.Len(t, m.Calls, 1, "expected exactly one interactor call")
		})
	}
}

// TestThreeCardRummyCuiController_BetForwardsBothStakes checks the parsed values
// reaching the interactor. Ante and Low Bonus differ, so a parser that swapped
// them, dropped the second argument, or reused the ante for both is caught.
func TestThreeCardRummyCuiController_BetForwardsBothStakes(t *testing.T) {
	tests := []struct {
		input        string
		wantAnte     int
		wantLowBonus int
	}{
		// No Low Bonus argument means no side bet, not a second ante.
		{"b 250", 250, 0},
		{"bet 250", 250, 0},
		{"b 250 75", 250, 75},
		{"bet 250 75", 250, 75},
		// 0 is a legal Low Bonus (the side bet is optional, min 0).
		{"b 250 0", 250, 0},
		// Extra whitespace is what strings.Fields exists to absorb.
		{"  bet   250   75  ", 250, 75},
	}

	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.input), func(t *testing.T) {
			m, c := newThreeCardRummyCuiFixture(t)
			assert.Equal(t, tcrCuiBetReply, c.Exec(tt.input))
			m.AssertCalled(t, "Bet", tt.wantAnte, tt.wantLowBonus)
		})
	}
}

// TestThreeCardRummyCuiController_BetRejectsBadInput pins that a malformed bet
// is answered with a rejection and never dispatched. The ante rejections are
// matched through i18n rather than against an English literal.
func TestThreeCardRummyCuiController_BetRejectsBadInput(t *testing.T) {
	t.Run("missing ante", func(t *testing.T) {
		m, c := newBareThreeCardRummyCuiController(t)
		out := c.Exec("b")
		assert.Equal(t, msgAnteAmountRequired(), out)
		assert.Empty(t, m.Calls, "a bet with no amount must not be dispatched")
	})

	t.Run("missing ante long form", func(t *testing.T) {
		m, c := newBareThreeCardRummyCuiController(t)
		assert.Equal(t, msgAnteAmountRequired(), c.Exec("bet"))
		assert.Empty(t, m.Calls)
	})

	for _, tt := range []struct{ name, input string }{
		{"not a number", "b abc"},
		{"zero", "b 0"},
		{"negative", "b -10"},
	} {
		t.Run("ante "+tt.name, func(t *testing.T) {
			m, c := newBareThreeCardRummyCuiController(t)
			out := c.Exec(tt.input)
			assert.Contains(t, out, msgInvalidAnteAmountPrefix())
			assert.True(t, msgRejected(out), "reply must carry the rejection marker: %q", out)
			assert.Empty(t, m.Calls, "an invalid ante must not be dispatched")
		})
	}

	// The Low Bonus is parsed with min 0, so a non-number or a negative amount
	// has to be refused too -- and refused *without* betting the ante, which is
	// the failure mode worth pinning: the ante parse has already succeeded by
	// the time the second argument is read.
	for _, tt := range []struct{ name, input string }{
		{"not a number", "b 100 abc"},
		{"negative", "b 100 -1"},
	} {
		t.Run("low bonus "+tt.name, func(t *testing.T) {
			m, c := newBareThreeCardRummyCuiController(t)
			out := c.Exec(tt.input)
			assert.True(t, msgRejected(out), "reply must carry the rejection marker: %q", out)
			assert.Empty(t, m.Calls, "an invalid low bonus must not bet the ante anyway")
		})
	}
}

// TestThreeCardRummyCuiController_UnknownCommand pins that an unrecognised verb
// is refused, names the offending input, and reaches no interactor method.
func TestThreeCardRummyCuiController_UnknownCommand(t *testing.T) {
	for _, cmd := range []string{"xyz", "pairplus", "double"} {
		t.Run(cmd, func(t *testing.T) {
			m, c := newBareThreeCardRummyCuiController(t)
			out := c.Exec(cmd)
			assert.True(t, msgRejected(out), "reply must carry the rejection marker: %q", out)
			assert.Contains(t, out, cmd, "the rejection should quote what was typed")
			assert.Empty(t, m.Calls, "an unknown command must not be dispatched")
		})
	}
}

// TestThreeCardRummyCuiController_KnownVerbsAreNotRejected is the negative
// control for the test above: every verb the controller advertises must be
// answered by the game, not by the unknown-command path.
func TestThreeCardRummyCuiController_KnownVerbsAreNotRejected(t *testing.T) {
	for _, input := range []string{"r", "reset", "b 100", "bet 100 50", "rb", "rebet", "p", "play", "f", "fold", "h", "hint", "log", "l"} {
		t.Run(input, func(t *testing.T) {
			_, c := newThreeCardRummyCuiFixture(t)
			out := c.Exec(input)
			assert.False(t, msgRejected(out), "%q must not be answered as an error: %q", input, out)
		})
	}
}

// TestThreeCardRummyCuiController_Quit covers the shared quit sentinel, which
// must not reach the interactor.
func TestThreeCardRummyCuiController_Quit(t *testing.T) {
	for _, cmd := range []string{"q", "quit", "exit"} {
		t.Run(cmd, func(t *testing.T) {
			m, c := newBareThreeCardRummyCuiController(t)
			assert.Equal(t, i18n.QuitSentinel, c.Exec(cmd))
			assert.Empty(t, m.Calls, "quit must not be dispatched")
		})
	}
}

// TestThreeCardRummyCuiController_EmptyInput pins that blank input prints the
// help hint rather than dispatching or crashing on fields[0].
func TestThreeCardRummyCuiController_EmptyInput(t *testing.T) {
	for _, input := range []string{"", "   ", "\t"} {
		m, c := newBareThreeCardRummyCuiController(t)
		assert.Equal(t, i18n.T("emptyInputHint"), c.Exec(input))
		assert.Empty(t, m.Calls, "empty input must not be dispatched")
	}
}

// TestThreeCardRummyCuiController_EveryMessageKeyResolves pins the i18n keys the
// controller hands to cuiutil.ParseIntArgKeys.
//
// **`i18n.T` returns the key itself when a translation is missing**, so nothing
// errors and no assertion built on `T(key)` can notice: `Contains(out, T(key))`
// compares the output against the key and passes in exactly the broken case.
// `invalidLowBonusAmount` shipped that way — the clone renamed the constant but
// not the locale entry, and a player who typed `b 100 abc` got the raw ASCII
// identifier back, in both languages. Assert the key *resolves* instead.
func TestThreeCardRummyCuiController_EveryMessageKeyResolves(t *testing.T) {
	keys := []string{"anteAmountRequired", "invalidAnteAmount", "invalidLowBonusAmount"}
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)
			for _, k := range keys {
				got := i18n.T(k)
				assert.NotEqual(t, k, got, "%s is missing from internal/i18n/locales/%s/common.json — "+
					"the CUI would print the raw identifier to the player", k, lang)
				assert.NotEmpty(t, got)
			}
		})
	}
}
