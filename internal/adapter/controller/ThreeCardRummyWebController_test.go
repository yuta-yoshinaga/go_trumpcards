//go:build test && (!js || !wasm || casino)

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// Every interactor method answers with a *distinct* payload, so an assertion on
// the response body pins which method the dispatcher reached. The Three Card
// Poker suite this game was cloned from hands the same payload back from every
// method, which means play/fold or hint/log could be swapped in the dispatcher
// and not one of its assertions would move.
const (
	tcrWebResetReply = `{"called":"reset"}`
	tcrWebBetReply   = `{"called":"bet"}`
	tcrWebRebetReply = `{"called":"rebet"}`
	tcrWebPlayReply  = `{"called":"play"}`
	tcrWebFoldReply  = `{"called":"fold"}`
	tcrWebHintReply  = `{"called":"hint"}`
	tcrWebLogReply   = `{"called":"actionLog"}`
)

// mustThreeCardRummyOutputJSON renders the default output the controller emits
// for the non-dispatched replies (quit, param errors).
func mustThreeCardRummyOutputJSON(msg string) string {
	out := &controller.ThreeCardRummyWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustThreeCardRummyOutputJSON: %v", err))
	}
	return string(b)
}

// newThreeCardRummyWebFixture builds a controller over a fully stubbed mock.
// The ante/lowBonus arguments are matched loosely here on purpose: the exact
// values are asserted by TestThreeCardRummyWebController_BetForwardsBothStakes,
// which would otherwise only see testify's "unexpected call" panic.
func newThreeCardRummyWebFixture(t *testing.T) (*usecase.MockThreeCardRummyInteractor, *controller.ThreeCardRummyWebController) {
	t.Helper()
	m := new(usecase.MockThreeCardRummyInteractor)
	m.On("Reset").Return(tcrWebResetReply)
	m.On("Bet", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(tcrWebBetReply)
	m.On("Rebet").Return(tcrWebRebetReply)
	m.On("Play").Return(tcrWebPlayReply)
	m.On("Fold").Return(tcrWebFoldReply)
	m.On("Hint").Return(tcrWebHintReply)
	m.On("ActionLog").Return(tcrWebLogReply)

	ctrl := controller.NewThreeCardRummyWebController(func() uc.ThreeCardRummyInteractorIF { return m })
	t.Cleanup(ctrl.Stop)
	return m, ctrl
}

// TestThreeCardRummyWebController_CommandRouting pins every command to the
// interactor method it must reach, by the distinct payload that method returns.
func TestThreeCardRummyWebController_CommandRouting(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		want   string
		method string
	}{
		{"reset", `{"command":"reset","sessionId":"tcr-reset"}`, tcrWebResetReply, "Reset"},
		{"reset alias r", `{"command":"r","sessionId":"tcr-reset-r"}`, tcrWebResetReply, "Reset"},
		{"bet", `{"command":"bet","amount":100,"sessionId":"tcr-bet"}`, tcrWebBetReply, "Bet"},
		{"bet alias b", `{"command":"b","amount":100,"sessionId":"tcr-bet-b"}`, tcrWebBetReply, "Bet"},
		{"bet with low bonus", `{"command":"bet","amount":100,"lowBonusBet":50,"sessionId":"tcr-bet-lb"}`, tcrWebBetReply, "Bet"},
		{"rebet", `{"command":"rebet","sessionId":"tcr-rebet"}`, tcrWebRebetReply, "Rebet"},
		{"rebet alias rb", `{"command":"rb","sessionId":"tcr-rebet-rb"}`, tcrWebRebetReply, "Rebet"},
		{"play", `{"command":"play","sessionId":"tcr-play"}`, tcrWebPlayReply, "Play"},
		{"play alias p", `{"command":"p","sessionId":"tcr-play-p"}`, tcrWebPlayReply, "Play"},
		{"fold", `{"command":"fold","sessionId":"tcr-fold"}`, tcrWebFoldReply, "Fold"},
		{"fold alias f", `{"command":"f","sessionId":"tcr-fold-f"}`, tcrWebFoldReply, "Fold"},
		{"hint", `{"command":"hint","sessionId":"tcr-hint"}`, tcrWebHintReply, "Hint"},
		{"hint alias h", `{"command":"h","sessionId":"tcr-hint-h"}`, tcrWebHintReply, "Hint"},
		{"log", `{"command":"log","sessionId":"tcr-log"}`, tcrWebLogReply, "ActionLog"},
		{"log alias l", `{"command":"l","sessionId":"tcr-log-l"}`, tcrWebLogReply, "ActionLog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newThreeCardRummyWebFixture(t)
			recorded := execRequest(t, ctrl.Exec, strings.NewReader(tt.body))
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(tt.want)
			m.AssertNumberOfCalls(t, tt.method, 1)
			if len(m.Calls) != 1 {
				t.Errorf("expected exactly one interactor call, got %d: %v", len(m.Calls), m.Calls)
			}
		})
	}
}

// TestThreeCardRummyWebController_BetForwardsBothStakes checks the values that
// reach the interactor, not merely that Bet was called. The two stakes differ so
// a dispatcher that swapped them, dropped one, or reused the ante for both is
// caught.
func TestThreeCardRummyWebController_BetForwardsBothStakes(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantAnte     int
		wantLowBonus int
	}{
		{
			name:         "both stakes",
			body:         `{"command":"bet","amount":250,"lowBonusBet":75,"sessionId":"tcr-stake-both"}`,
			wantAnte:     250,
			wantLowBonus: 75,
		},
		{
			// The Low Bonus is optional, so an absent field must arrive as 0
			// rather than as a copy of the ante.
			name:         "low bonus omitted",
			body:         `{"command":"b","amount":250,"sessionId":"tcr-stake-ante"}`,
			wantAnte:     250,
			wantLowBonus: 0,
		},
		{
			name:         "low bonus explicitly null",
			body:         `{"command":"bet","amount":40,"lowBonusBet":null,"sessionId":"tcr-stake-null"}`,
			wantAnte:     40,
			wantLowBonus: 0,
		},
		{
			name:         "low bonus explicitly zero",
			body:         `{"command":"bet","amount":40,"lowBonusBet":0,"sessionId":"tcr-stake-zero"}`,
			wantAnte:     40,
			wantLowBonus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newThreeCardRummyWebFixture(t)
			recorded := execRequest(t, ctrl.Exec, strings.NewReader(tt.body))
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(tcrWebBetReply)
			m.AssertCalled(t, "Bet", tt.wantAnte, tt.wantLowBonus)
		})
	}
}

// TestThreeCardRummyWebController_UnknownCommand pins the default branch: an
// unrecognised command must answer 400 with the unsupported-command output and
// must not reach the interactor at all. The mock carries no expectations, so any
// dispatch would blow up rather than pass silently.
func TestThreeCardRummyWebController_UnknownCommand(t *testing.T) {
	for _, cmd := range []string{"xyz", "pairplus", "s", "n"} {
		t.Run(cmd, func(t *testing.T) {
			m := new(usecase.MockThreeCardRummyInteractor)
			ctrl := controller.NewThreeCardRummyWebController(func() uc.ThreeCardRummyInteractorIF { return m })
			t.Cleanup(ctrl.Stop)

			body := fmt.Sprintf(`{"command":%q,"sessionId":"tcr-unknown-%s"}`, cmd, cmd)
			recorded := execRequest(t, ctrl.Exec, strings.NewReader(body))
			recorded.CodeIs(http.StatusBadRequest)
			recorded.BodyIs(mustThreeCardRummyOutputJSON("Unsupported command."))
			if len(m.Calls) != 0 {
				t.Errorf("unknown command %q reached the interactor: %v", cmd, m.Calls)
			}
		})
	}
}

// TestThreeCardRummyWebController_QuitAndParamErrors covers the replies that
// never reach the dispatcher.
func TestThreeCardRummyWebController_QuitAndParamErrors(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		for _, cmd := range []string{"q", "quit"} {
			m, ctrl := newThreeCardRummyWebFixture(t)
			body := fmt.Sprintf(`{"command":%q,"sessionId":"tcr-quit"}`, cmd)
			recorded := execRequest(t, ctrl.Exec, strings.NewReader(body))
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mustThreeCardRummyOutputJSON("bye."))
			if len(m.Calls) != 0 {
				t.Errorf("%q reached the interactor: %v", cmd, m.Calls)
			}
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		m, ctrl := newThreeCardRummyWebFixture(t)
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustThreeCardRummyOutputJSON("param error."))
		if len(m.Calls) != 0 {
			t.Errorf("malformed request reached the interactor: %v", m.Calls)
		}
	})

	t.Run("missing session id", func(t *testing.T) {
		m, ctrl := newThreeCardRummyWebFixture(t)
		recorded := execRequest(t, ctrl.Exec, strings.NewReader(`{"command":"play"}`))
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustThreeCardRummyOutputJSON("param error."))
		if len(m.Calls) != 0 {
			t.Errorf("session-less request reached the interactor: %v", m.Calls)
		}
	})

	t.Run("missing command", func(t *testing.T) {
		m, ctrl := newThreeCardRummyWebFixture(t)
		recorded := execRequest(t, ctrl.Exec, strings.NewReader(`{"sessionId":"tcr-nocmd"}`))
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustThreeCardRummyOutputJSON("param error."))
		if len(m.Calls) != 0 {
			t.Errorf("command-less request reached the interactor: %v", m.Calls)
		}
	})
}

// TestThreeCardRummyWebInput_DeclaresOnlyFieldsACommandReads guards against the
// clone leftovers this file exists to catch. Three Card Rummy was cloned from
// Three Card Poker, whose input carries PairPlusBet; a field left behind after a
// rename is accepted off the wire, documented by its json tag, and then silently
// ignored — no compiler or lint error says so.
//
// The two names below are exactly the fields
// TestThreeCardRummyWebController_BetForwardsBothStakes proves are forwarded.
// Adding a field here means proving a command reads it there.
func TestThreeCardRummyWebInput_DeclaresOnlyFieldsACommandReads(t *testing.T) {
	readByACommand := map[string]bool{
		"Amount":      false, // "b"/"bet" -> ante
		"LowBonusBet": false, // "b"/"bet" -> low bonus side bet
	}

	typ := reflect.TypeOf(controller.ThreeCardRummyWebInput{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		// BaseWebInput carries command/sessionId/n for every game; it is not a
		// per-game field and is not this test's business.
		if f.Anonymous {
			continue
		}
		if _, ok := readByACommand[f.Name]; !ok {
			t.Errorf("ThreeCardRummyWebInput.%s (json %q) is declared but no command reads it -- leftover from the Three Card Poker clone?",
				f.Name, f.Tag.Get("json"))
			continue
		}
		readByACommand[f.Name] = true
	}

	for name, seen := range readByACommand {
		if !seen {
			t.Errorf("ThreeCardRummyWebInput.%s is expected by the dispatcher tests but the struct no longer declares it", name)
		}
	}
}
