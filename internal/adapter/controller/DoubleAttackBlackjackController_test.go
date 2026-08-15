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
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestDoubleAttackWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.DoubleAttackBlackjackWebOutput{
		Hands:       make([]*controller.DoubleAttackWebOutputHand, 0),
		DealerCards: make([]*controller.WebOutputCard, 0),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"hands", "dealerCards"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func newDAMock() *usecase.MockDoubleAttackBlackjackInteractor {
	m := new(usecase.MockDoubleAttackBlackjackInteractor)
	out := `{"phase":0,"hands":[],"dealerCards":[],"chips":1000,"message":""}`
	m.On("Reset").Return(out)
	m.On("PlaceBet", 50, 20).Return(out)
	m.On("PlaceBet", 50, 0).Return(out)
	m.On("Attack", 50).Return(out)
	m.On("Attack", 0).Return(out)
	m.On("Hit").Return(out)
	m.On("Stand").Return(out)
	m.On("Double").Return(out)
	m.On("Split").Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

func TestDoubleAttackWebController_Method(t *testing.T) {
	m := newDAMock()
	ctrl := controller.NewDoubleAttackBlackjackWebController(
		func() uc.DoubleAttackBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"bet", `{"command":"bet","ante":50,"bustIt":20,"sessionId":"s2"}`},
		{"bet without side", `{"command":"bet","ante":50,"sessionId":"s3"}`},
		{"attack", `{"command":"attack","amount":50,"sessionId":"s4"}`},
		{"attack declined", `{"command":"attack","amount":0,"sessionId":"s5"}`},
		{"hit", `{"command":"hit","sessionId":"s6"}`},
		{"stand", `{"command":"stand","sessionId":"s7"}`},
		{"double", `{"command":"double","sessionId":"s8"}`},
		{"split", `{"command":"split","sessionId":"s9"}`},
		{"next", `{"command":"next","sessionId":"s10"}`},
		{"hint", `{"command":"hint","sessionId":"s11"}`},
		{"log", `{"command":"log","sessionId":"s12"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.DoubleAttackBlackjackWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.DoubleAttackBlackjackWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **追加ベットの見送り (0) は「送らなかった」とは違う。**
func TestDoubleAttackWebController_ZeroAttackIsAValue(t *testing.T) {
	m := newDAMock()
	ctrl := controller.NewDoubleAttackBlackjackWebController(
		func() uc.DoubleAttackBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	var declined controller.DoubleAttackBlackjackWebInput
	if err := json.Unmarshal([]byte(`{"command":"attack","amount":0,"sessionId":"z1"}`), &declined); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &declined).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Attack", 0)

	var missing controller.DoubleAttackBlackjackWebInput
	if err := json.Unmarshal([]byte(`{"command":"attack","sessionId":"z2"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "amount is required") {
		t.Errorf("expected 'amount is required', got: %s", recorded.Body.String())
	}
	m.AssertNumberOfCalls(t, "Attack", 1)
}

// **上限はコントローラで丸めない。** アンティ次第なのでドメインへそのまま渡す。
func TestDoubleAttackWebController_AttackIsNotClampedHere(t *testing.T) {
	m := new(usecase.MockDoubleAttackBlackjackInteractor)
	m.On("Attack", 99999).Return(`{"message":"too much"}`)
	ctrl := controller.NewDoubleAttackBlackjackWebController(
		func() uc.DoubleAttackBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.DoubleAttackBlackjackWebInput
	if err := json.Unmarshal([]byte(`{"command":"attack","amount":99999,"sessionId":"c1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Attack", 99999)
}

func TestDoubleAttackWebController_AnteRequired(t *testing.T) {
	m := newDAMock()
	ctrl := controller.NewDoubleAttackBlackjackWebController(
		func() uc.DoubleAttackBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.DoubleAttackBlackjackWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"a1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "ante is required") {
		t.Errorf("expected 'ante is required', got: %s", recorded.Body.String())
	}
}

// --- CUI ---

func TestDoubleAttackCuiController(t *testing.T) {
	m := newDAMock()
	c := controller.NewDoubleAttackBlackjackCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("bet 50 20"))
	assert.NotEmpty(t, c.Exec("bet 50"))
	m.AssertCalled(t, "PlaceBet", 50, 0)
	assert.True(t, msgRejected(c.Exec("bet")))
	assert.True(t, msgRejected(c.Exec("bet xyz")))

	// **見送り (0) が通ること。**
	assert.NotEmpty(t, c.Exec("attack 0"))
	m.AssertCalled(t, "Attack", 0)
	assert.True(t, msgRejected(c.Exec("attack")))

	for _, cmd := range []string{"hit", "stand", "double", "split", "next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
