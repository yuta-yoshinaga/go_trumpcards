//go:build test && (!js || !wasm || extra)

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// mockSpeculationInteractor is a local testify mock for
// usecase.SpeculationInteractorIF.
//
// The shared mock at internal/adapter/controller/usecase/
// SpeculationInteractor_mock.go is a stale clone of a betting game: it has
// ResetWithConfig/PlaceBet/GetConfig and is missing Flip/Accept/Decline/Bid,
// so it does not satisfy SpeculationInteractorIF at all.
type mockSpeculationInteractor struct {
	mock.Mock
}

func (m *mockSpeculationInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

func (m *mockSpeculationInteractor) Reset() string         { return m.Called().String(0) }
func (m *mockSpeculationInteractor) Flip() string          { return m.Called().String(0) }
func (m *mockSpeculationInteractor) Accept() string        { return m.Called().String(0) }
func (m *mockSpeculationInteractor) Decline() string       { return m.Called().String(0) }
func (m *mockSpeculationInteractor) Bid(amount int) string { return m.Called(amount).String(0) }
func (m *mockSpeculationInteractor) NextRound() string     { return m.Called().String(0) }
func (m *mockSpeculationInteractor) Hint() string          { return m.Called().String(0) }
func (m *mockSpeculationInteractor) ActionLog() string     { return m.Called().String(0) }

var _ uc.SpeculationInteractorIF = (*mockSpeculationInteractor)(nil)

// newMockSpeculationInteractor registers a distinct return string per verb, so
// a dispatch that routes a command to the wrong method is visible in the body
// rather than hidden behind a shared "ok".
func newMockSpeculationInteractor() *mockSpeculationInteractor {
	m := new(mockSpeculationInteractor)
	m.On("Reset").Return("reset result")
	m.On("Flip").Return("flip result")
	m.On("Accept").Return("accept result")
	m.On("Decline").Return("decline result")
	m.On("Bid", 50).Return("bid 50 result")
	m.On("Bid", 7).Return("bid 7 result")
	m.On("NextRound").Return("next result")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("log result")
	return m
}

// --- CUI ---

func TestSpeculationCuiController_Quit(t *testing.T) {
	c := controller.NewSpeculationCuiController(newMockSpeculationInteractor())
	assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
	assert.Equal(t, i18n.QuitSentinel, c.Exec("quit"))
}

func TestSpeculationCuiController_Reset(t *testing.T) {
	c := controller.NewSpeculationCuiController(newMockSpeculationInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestSpeculationCuiController_Flip(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "flip result", c.Exec("f"))
	assert.Equal(t, "flip result", c.Exec("flip"))
	m.AssertNumberOfCalls(t, "Flip", 2)
}

func TestSpeculationCuiController_Accept(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "accept result", c.Exec("a"))
	assert.Equal(t, "accept result", c.Exec("accept"))
	m.AssertNumberOfCalls(t, "Accept", 2)
	// Accepting must never be routed to the opposite answer.
	m.AssertNotCalled(t, "Decline")
}

func TestSpeculationCuiController_Decline(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "decline result", c.Exec("d"))
	assert.Equal(t, "decline result", c.Exec("decline"))
	m.AssertNumberOfCalls(t, "Decline", 2)
	m.AssertNotCalled(t, "Accept")
}

func TestSpeculationCuiController_Bid(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "bid 50 result", c.Exec("bid 50"))
	assert.Equal(t, "bid 7 result", c.Exec("bid 7"))
	m.AssertCalled(t, "Bid", 50)
	m.AssertCalled(t, "Bid", 7)
}

// TestSpeculationCuiController_Bid_Rejected pins that a bid the parser cannot
// read never reaches the interactor.
//
// **A missing or unreadable amount must not become 0.** Bidding 0 is not the
// same move as declining, and a silent 0 would let the player "buy" the lead
// for nothing.
func TestSpeculationCuiController_Bid_Rejected(t *testing.T) {
	cases := []struct{ name, cmd string }{
		{"missing amount", "bid"},
		{"non-numeric", "bid abc"},
		{"zero", "bid 0"},
		{"negative", "bid -5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newMockSpeculationInteractor()
			c := controller.NewSpeculationCuiController(m)
			out := c.Exec(tc.cmd)

			assert.NotEmpty(t, out)
			// The reply is a rejection, not a board.
			assert.True(t, strings.HasPrefix(out, i18n.ErrorPrefix),
				"expected a marked rejection, got %q", out)
			assert.NotContains(t, out, "bid 50 result")
			// Nothing was dispatched: no amount at all reached the domain.
			m.AssertNotCalled(t, "Bid", mock.Anything)
		})
	}
}

// TestSpeculationCuiController_Bid_RejectionIsTranslated guards the reverse of
// the vacuous i18n check: i18n.T returns the key when a translation is missing,
// so assert the real text rather than T(key) == T(key).
func TestSpeculationCuiController_Bid_RejectionIsTranslated(t *testing.T) {
	c := controller.NewSpeculationCuiController(newMockSpeculationInteractor())
	assert.Contains(t, c.Exec("bid"), "ベットを指定してください。")
	assert.Contains(t, c.Exec("bid abc"), "ベットが不正です。")
}

func TestSpeculationCuiController_Next(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "next result", c.Exec("next"))
	m.AssertCalled(t, "NextRound")
}

func TestSpeculationCuiController_Hint(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	// Both spellings: the help text advertises "h / hint", and a verb that is
	// advertised but not accepted is worse than one that was never offered.
	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	m.AssertNumberOfCalls(t, "Hint", 2)
}

func TestSpeculationCuiController_ActionLog(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	assert.Equal(t, "log result", c.Exec("log"))
	assert.Equal(t, "log result", c.Exec("l"))
	m.AssertNumberOfCalls(t, "ActionLog", 2)
}

func TestSpeculationCuiController_Unknown(t *testing.T) {
	m := newMockSpeculationInteractor()
	c := controller.NewSpeculationCuiController(m)
	out := c.Exec("xyz")
	assert.Contains(t, out, "コマンドが不明です")
	assert.Contains(t, out, "xyz")
	m.AssertNotCalled(t, "Flip")
	m.AssertNotCalled(t, "Reset")
}

func TestSpeculationCuiController_Empty(t *testing.T) {
	c := controller.NewSpeculationCuiController(newMockSpeculationInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

// --- Web ---

func mustSpeculationOutputJSON(msg string) string {
	out := &controller.SpeculationWebOutput{
		Seats:         make([]*controller.SpeculationWebOutputSeat, 0),
		TrumpSuit:     -1,
		BestSeat:      -1,
		OfferFrom:     -1,
		OfferTo:       -1,
		WinnerSeat:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSpeculationOutputJSON: %v", err))
	}
	return string(b)
}

func newSpeculationWebController(m *mockSpeculationInteractor) *controller.SpeculationWebController {
	return controller.NewSpeculationWebController(func() uc.SpeculationInteractorIF { return m })
}

func TestSpeculationWebController_Dispatch(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	cases := []struct{ name, body, want string }{
		{"flip short", `{"command":"f","sessionId":"w1"}`, "flip result"},
		{"flip long", `{"command":"flip","sessionId":"w2"}`, "flip result"},
		{"accept short", `{"command":"a","sessionId":"w3"}`, "accept result"},
		{"accept long", `{"command":"accept","sessionId":"w4"}`, "accept result"},
		{"decline short", `{"command":"d","sessionId":"w5"}`, "decline result"},
		{"decline long", `{"command":"decline","sessionId":"w6"}`, "decline result"},
		{"bid", `{"command":"bid","amount":50,"sessionId":"w7"}`, "bid 50 result"},
		{"next", `{"command":"next","sessionId":"w8"}`, "next result"},
		{"hint", `{"command":"hint","sessionId":"w9"}`, "hint result"},
		{"reset short", `{"command":"r","sessionId":"w10"}`, "reset result"},
		{"reset long", `{"command":"reset","sessionId":"w11"}`, "reset result"},
		{"log", `{"command":"log","sessionId":"w12"}`, "log result"},
		{"log short", `{"command":"l","sessionId":"w13"}`, "log result"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.SpeculationWebInput
			if err := json.Unmarshal([]byte(tc.body), &input); err != nil {
				t.Fatalf("bad fixture: %v", err)
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(tc.want)
		})
	}
}

func TestSpeculationWebController_BidForwardsAmount(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	var input controller.SpeculationWebInput
	if err := json.Unmarshal([]byte(`{"command":"bid","amount":7,"sessionId":"wb"}`), &input); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs("bid 7 result")
	m.AssertCalled(t, "Bid", 7)
	m.AssertNotCalled(t, "Bid", 50)
}

// TestSpeculationWebController_BidRequiresAmount pins that an omitted amount is
// a 400, not a silent Bid(0). Asserting only the status code would pass even if
// the interactor had already been called, so this also asserts nothing was
// dispatched.
func TestSpeculationWebController_BidRequiresAmount(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	var input controller.SpeculationWebInput
	if err := json.Unmarshal([]byte(`{"command":"bid","sessionId":"wna"}`), &input); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)

	body := strings.TrimSpace(recorded.Body.String())
	assert.Contains(t, body, "amount is required")
	// The error body keeps the game's output shape, sentinels and all.
	assert.Contains(t, body, `"bestSeat":-1`)

	m.AssertNotCalled(t, "Bid", mock.Anything)
	m.AssertNotCalled(t, "Bid", 0)
}

func TestSpeculationWebController_Quit(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	var input controller.SpeculationWebInput
	if err := json.Unmarshal([]byte(`{"command":"q","sessionId":"wq"}`), &input); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mustSpeculationOutputJSON("bye."))
}

func TestSpeculationWebController_UnknownCommand(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	var input controller.SpeculationWebInput
	if err := json.Unmarshal([]byte(`{"command":"xyz","sessionId":"wu"}`), &input); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	assert.Contains(t, recorded.Body.String(), "Unsupported command")
	m.AssertNotCalled(t, "Reset")
	m.AssertNotCalled(t, "Flip")
}

func TestSpeculationWebController_InvalidJSON(t *testing.T) {
	m := newMockSpeculationInteractor()
	ctrl := newSpeculationWebController(m)
	defer ctrl.Stop()

	recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
	recorded.CodeIs(http.StatusBadRequest)
}

// TestSpeculationWebController_DefaultOutputSentinels pins that "no seat" is
// -1 everywhere. **0 is a valid seat** (it is the human), so a 0 default would
// render an empty table as "you hold the best trump".
func TestSpeculationWebController_DefaultOutputSentinels(t *testing.T) {
	var out controller.SpeculationWebOutput
	if err := json.Unmarshal([]byte(mustSpeculationOutputJSON("bye.")), &out); err != nil {
		t.Fatalf("default output is not valid JSON: %v", err)
	}
	assert.Equal(t, -1, out.BestSeat)
	assert.Equal(t, -1, out.OfferFrom)
	assert.Equal(t, -1, out.OfferTo)
	assert.Equal(t, -1, out.WinnerSeat)
	assert.Equal(t, -1, out.TrumpSuit)
	assert.NotNil(t, out.Seats)
}
