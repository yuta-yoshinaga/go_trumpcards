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
func TestFreeBetWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.FreeBetBlackjackWebOutput{
		Hands:       make([]*controller.FreeBetWebOutputHand, 0),
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

// **プレイヤーの金とハウスの金は別々の欄で出る。**
func TestFreeBetWebController_HandCarriesBothMoneySources(t *testing.T) {
	b, err := json.Marshal(&controller.FreeBetWebOutputHand{Bet: 50, FreeBet: 50})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(raw["bet"]) != "50" || string(raw["freeBet"]) != "50" {
		t.Errorf("bet=%s freeBet=%s, want 50/50", raw["bet"], raw["freeBet"])
	}
}

func newFBMock() *usecase.MockFreeBetBlackjackInteractor {
	m := new(usecase.MockFreeBetBlackjackInteractor)
	out := `{"phase":0,"hands":[],"dealerCards":[],"chips":1000,"message":""}`
	m.On("Reset").Return(out)
	m.On("PlaceBet", 50).Return(out)
	m.On("Hit").Return(out)
	m.On("Stand").Return(out)
	m.On("FreeDouble").Return(out)
	m.On("FreeSplit").Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

func TestFreeBetWebController_Method(t *testing.T) {
	m := newFBMock()
	ctrl := controller.NewFreeBetBlackjackWebController(
		func() uc.FreeBetBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"bet", `{"command":"bet","ante":50,"sessionId":"s2"}`},
		{"bet alias", `{"command":"b","ante":50,"sessionId":"s3"}`},
		{"hit", `{"command":"hit","sessionId":"s4"}`},
		{"stand", `{"command":"stand","sessionId":"s5"}`},
		{"free double", `{"command":"freedouble","sessionId":"s6"}`},
		{"free double alias", `{"command":"fd","sessionId":"s7"}`},
		{"free split", `{"command":"freesplit","sessionId":"s8"}`},
		{"free split alias", `{"command":"fs","sessionId":"s9"}`},
		{"next", `{"command":"next","sessionId":"s10"}`},
		{"hint", `{"command":"hint","sessionId":"s11"}`},
		{"log", `{"command":"log","sessionId":"s12"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.FreeBetBlackjackWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.FreeBetBlackjackWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

func TestFreeBetWebController_AnteRequired(t *testing.T) {
	m := newFBMock()
	ctrl := controller.NewFreeBetBlackjackWebController(
		func() uc.FreeBetBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.FreeBetBlackjackWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"a1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "ante is required") {
		t.Errorf("expected 'ante is required', got: %s", recorded.Body.String())
	}
	m.AssertNumberOfCalls(t, "PlaceBet", 0)
}

// **無料操作はコントローラで判定しない。** 可否はドメインだけが知っている。
func TestFreeBetWebController_FreeActionsAreNotGatedHere(t *testing.T) {
	m := new(usecase.MockFreeBetBlackjackInteractor)
	m.On("FreeDouble").Return(`{"message":"cannot free double now"}`)
	m.On("FreeSplit").Return(`{"message":"cannot free split now"}`)
	ctrl := controller.NewFreeBetBlackjackWebController(
		func() uc.FreeBetBlackjackInteractorIF { return m })
	defer ctrl.Stop()

	for _, cmd := range []string{"fd", "fs"} {
		var input controller.FreeBetBlackjackWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"g1"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	}
	m.AssertCalled(t, "FreeDouble")
	m.AssertCalled(t, "FreeSplit")
}

// --- CUI ---

func TestFreeBetCuiController(t *testing.T) {
	m := newFBMock()
	c := controller.NewFreeBetBlackjackCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("bet 50"))
	m.AssertCalled(t, "PlaceBet", 50)
	assert.Contains(t, c.Exec("bet"), "required")
	assert.Contains(t, c.Exec("bet xyz"), "Invalid")

	for _, cmd := range []string{"hit", "h", "stand", "s", "fd", "freedouble",
		"fs", "freesplit", "next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	m.AssertNumberOfCalls(t, "FreeDouble", 2)
	m.AssertNumberOfCalls(t, "FreeSplit", 2)
	assert.NotEmpty(t, c.Exec("zzz"))
}
