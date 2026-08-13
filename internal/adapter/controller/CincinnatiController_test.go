//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestCincinnatiWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.CincinnatiWebOutput{
		Seats:     make([]*controller.CincinnatiWebOutputSeat, 0),
		Community: make([]*controller.WebOutputCard, 0),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"seats", "community"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func newCinMock() *usecase.MockCincinnatiInteractor {
	m := new(usecase.MockCincinnatiInteractor)
	out := `{"phase":1,"seats":[],"community":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("Action", domain.CincinnatiActionFold, 0).Return(out)
	m.On("Action", domain.CincinnatiActionCheck, 0).Return(out)
	m.On("Action", domain.CincinnatiActionCall, 0).Return(out)
	m.On("Action", domain.CincinnatiActionBet, 20).Return(out)
	m.On("Action", domain.CincinnatiActionRaise, 20).Return(out)
	m.On("NextHand").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

// **コマンド名がドメインのアクション値に正しく対応する。**
//
// ここが 1 つずれると「チェックしたのに降りる」形で静かに壊れる。
func TestCincinnatiWebController_MapsCommandsToActions(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantAction int
		wantAmount int
	}{
		{"fold", `{"command":"fold","sessionId":"s1"}`, domain.CincinnatiActionFold, 0},
		{"fold alias", `{"command":"f","sessionId":"s2"}`, domain.CincinnatiActionFold, 0},
		{"check", `{"command":"check","sessionId":"s3"}`, domain.CincinnatiActionCheck, 0},
		{"check alias", `{"command":"k","sessionId":"s4"}`, domain.CincinnatiActionCheck, 0},
		{"call", `{"command":"call","sessionId":"s5"}`, domain.CincinnatiActionCall, 0},
		{"call alias", `{"command":"c","sessionId":"s6"}`, domain.CincinnatiActionCall, 0},
		{"bet", `{"command":"bet","amount":20,"sessionId":"s7"}`, domain.CincinnatiActionBet, 20},
		{"raise", `{"command":"raise","amount":20,"sessionId":"s8"}`, domain.CincinnatiActionRaise, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newCinMock()
			ctrl := controller.NewCincinnatiWebController(func() uc.CincinnatiInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.CincinnatiWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "Action", tc.wantAction, tc.wantAmount)
		})
	}
}

func TestCincinnatiWebController_Method(t *testing.T) {
	m := newCinMock()
	ctrl := controller.NewCincinnatiWebController(func() uc.CincinnatiInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"m1"}`},
		{"next", `{"command":"next","sessionId":"m2"}`},
		{"hint", `{"command":"hint","sessionId":"m3"}`},
		{"log", `{"command":"log","sessionId":"m4"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.CincinnatiWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.CincinnatiWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"m9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **額が要る手では未送信を弾く。** 額の要らない手では送られても無視する。
func TestCincinnatiWebController_AmountIsRequiredOnlyWhereItMatters(t *testing.T) {
	m := newCinMock()
	ctrl := controller.NewCincinnatiWebController(func() uc.CincinnatiInteractorIF { return m })
	defer ctrl.Stop()

	for _, cmd := range []string{"bet", "raise"} {
		var input controller.CincinnatiWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a1"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "amount is required") {
			t.Errorf("%s: expected 'amount is required', got: %s", cmd, recorded.Body.String())
		}
	}

	// 降りる・チェック・コールは額を送らなくても通る。
	for _, cmd := range []string{"fold", "check", "call"} {
		var input controller.CincinnatiWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a2"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	}
}

// --- CUI ---

func TestCincinnatiCuiController(t *testing.T) {
	m := newCinMock()
	c := controller.NewCincinnatiCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("fold"))
	m.AssertCalled(t, "Action", domain.CincinnatiActionFold, 0)
	assert.NotEmpty(t, c.Exec("k"))
	m.AssertCalled(t, "Action", domain.CincinnatiActionCheck, 0)
	assert.NotEmpty(t, c.Exec("call"))
	m.AssertCalled(t, "Action", domain.CincinnatiActionCall, 0)

	assert.NotEmpty(t, c.Exec("bet 20"))
	m.AssertCalled(t, "Action", domain.CincinnatiActionBet, 20)
	assert.NotEmpty(t, c.Exec("raise 20"))
	m.AssertCalled(t, "Action", domain.CincinnatiActionRaise, 20)

	assert.Contains(t, c.Exec("bet"), "required")
	assert.Contains(t, c.Exec("bet xyz"), "Invalid")

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
