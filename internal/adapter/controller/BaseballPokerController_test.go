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
func TestBaseballPokerWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.BaseballPokerWebOutput{
		Seats:      make([]*controller.BaseballPokerWebOutputSeat, 0),
		WildValues: []int{domain.BaseballWildThree, domain.BaseballWildNine},
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"seats", "wildValues"} {
		if got := string(raw[key]); got == "null" {
			t.Errorf("%s = null, want an array", key)
		}
	}
}

func newBbMock() *usecase.MockBaseballPokerInteractor {
	m := new(usecase.MockBaseballPokerInteractor)
	out := `{"phase":0,"seats":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("Action", domain.BaseballActionFold, 0).Return(out)
	m.On("Action", domain.BaseballActionCheck, 0).Return(out)
	m.On("Action", domain.BaseballActionCall, 0).Return(out)
	m.On("Action", domain.BaseballActionBet, 20).Return(out)
	m.On("Action", domain.BaseballActionRaise, 20).Return(out)
	m.On("AnswerBuyIn", domain.BaseballBuyPay).Return(out)
	m.On("AnswerBuyIn", domain.BaseballBuyFold).Return(out)
	m.On("NextHand").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

// **コマンド名がドメインのアクション値に正しく対応する。**
func TestBaseballPokerWebController_MapsCommandsToActions(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantAction int
		wantAmount int
	}{
		{"fold", `{"command":"fold","sessionId":"s1"}`, domain.BaseballActionFold, 0},
		{"fold alias", `{"command":"f","sessionId":"s2"}`, domain.BaseballActionFold, 0},
		{"check", `{"command":"check","sessionId":"s3"}`, domain.BaseballActionCheck, 0},
		{"check alias", `{"command":"k","sessionId":"s4"}`, domain.BaseballActionCheck, 0},
		{"call", `{"command":"call","sessionId":"s5"}`, domain.BaseballActionCall, 0},
		{"call alias", `{"command":"c","sessionId":"s6"}`, domain.BaseballActionCall, 0},
		{"bet", `{"command":"bet","amount":20,"sessionId":"s7"}`, domain.BaseballActionBet, 20},
		{"raise", `{"command":"raise","amount":20,"sessionId":"s8"}`, domain.BaseballActionRaise, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newBbMock()
			ctrl := controller.NewBaseballPokerWebController(func() uc.BaseballPokerInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.BaseballPokerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "Action", tc.wantAction, tc.wantAmount)
		})
	}
}

// **買い増しの返事は名前で受ける。** 数値の本文にすると、送り忘れが
// 「0 番の返事」= 支払いに化けて、降りたつもりの席がポットを払う。
func TestBaseballPokerWebController_MapsBuyInAnswers(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantAnswer int
	}{
		{"pay", `{"command":"pay","sessionId":"b1"}`, domain.BaseballBuyPay},
		{"pay alias", `{"command":"p","sessionId":"b2"}`, domain.BaseballBuyPay},
		{"fold instead", `{"command":"buyfold","sessionId":"b3"}`, domain.BaseballBuyFold},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newBbMock()
			ctrl := controller.NewBaseballPokerWebController(func() uc.BaseballPokerInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.BaseballPokerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "AnswerBuyIn", tc.wantAnswer)
		})
	}

	// 降りるつもりのコマンドが支払いに化けていないことを、逆向きにも見る。
	m := newBbMock()
	ctrl := controller.NewBaseballPokerWebController(func() uc.BaseballPokerInteractorIF { return m })
	defer ctrl.Stop()
	var input controller.BaseballPokerWebInput
	_ = json.Unmarshal([]byte(`{"command":"buyfold","sessionId":"b9"}`), &input)
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertNotCalled(t, "AnswerBuyIn", domain.BaseballBuyPay)
}

func TestBaseballPokerWebController_Method(t *testing.T) {
	m := newBbMock()
	ctrl := controller.NewBaseballPokerWebController(func() uc.BaseballPokerInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"m1"}`},
		{"next", `{"command":"next","sessionId":"m2"}`},
		{"hint", `{"command":"hint","sessionId":"m3"}`},
		{"log", `{"command":"log","sessionId":"m4"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.BaseballPokerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.BaseballPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"m9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **額が要る手では未送信を弾く。** 額の要らない手では送られても無視する。
func TestBaseballPokerWebController_AmountIsRequiredOnlyWhereItMatters(t *testing.T) {
	m := newBbMock()
	ctrl := controller.NewBaseballPokerWebController(func() uc.BaseballPokerInteractorIF { return m })
	defer ctrl.Stop()

	for _, cmd := range []string{"bet", "raise"} {
		var input controller.BaseballPokerWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a1"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "amount is required") {
			t.Errorf("%s: expected 'amount is required', got: %s", cmd, recorded.Body.String())
		}
	}

	for _, cmd := range []string{"fold", "check", "call", "pay", "buyfold"} {
		var input controller.BaseballPokerWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a2"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	}
}

// --- CUI ---

func TestBaseballPokerCuiController(t *testing.T) {
	m := newBbMock()
	c := controller.NewBaseballPokerCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("fold"))
	m.AssertCalled(t, "Action", domain.BaseballActionFold, 0)
	assert.NotEmpty(t, c.Exec("k"))
	m.AssertCalled(t, "Action", domain.BaseballActionCheck, 0)
	assert.NotEmpty(t, c.Exec("call"))
	m.AssertCalled(t, "Action", domain.BaseballActionCall, 0)

	assert.NotEmpty(t, c.Exec("bet 20"))
	m.AssertCalled(t, "Action", domain.BaseballActionBet, 20)
	assert.NotEmpty(t, c.Exec("raise 20"))
	m.AssertCalled(t, "Action", domain.BaseballActionRaise, 20)

	assert.NotEmpty(t, c.Exec("pay"))
	m.AssertCalled(t, "AnswerBuyIn", domain.BaseballBuyPay)
	assert.NotEmpty(t, c.Exec("buyfold"))
	m.AssertCalled(t, "AnswerBuyIn", domain.BaseballBuyFold)

	assert.Contains(t, c.Exec("bet"), msgAmountRequired())
	assert.Contains(t, c.Exec("bet xyz"), msgInvalidAmountPrefix())

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
