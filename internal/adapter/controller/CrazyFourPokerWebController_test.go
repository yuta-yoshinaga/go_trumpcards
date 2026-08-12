//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestCrazyFourPokerWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.CrazyFourPokerWebOutput{
		PlayerHand: make([]*controller.WebOutputCard, 0),
		DealerHand: make([]*controller.WebOutputCard, 0),
		PlayerBest: make([]*controller.WebOutputCard, 0),
		DealerBest: make([]*controller.WebOutputCard, 0),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"playerHand", "dealerHand", "playerBest", "dealerBest"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func TestCrazyFourPokerWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0,"playerHand":[],"dealerHand":[],"playerBest":[],"dealerBest":[],` +
		`"chips":1000,"roundNumber":1,"gameEndFlag":false,"message":""}`

	m := new(usecase.MockCrazyFourPokerInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("PlaceBet", 50, 20).Return(mockOutput)
	m.On("PlaceBet", 50, 0).Return(mockOutput)
	m.On("Play", 3).Return(mockOutput)
	m.On("Fold").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	factory := func() uc.CrazyFourPokerInteractorIF { return m }
	ctrl := controller.NewCrazyFourPokerWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"bet", `{"command":"bet","ante":50,"queensUp":20,"sessionId":"s2"}`},
		{"bet without side", `{"command":"bet","ante":50,"sessionId":"s3"}`},
		{"play", `{"command":"play","multiplier":3,"sessionId":"s4"}`},
		{"fold", `{"command":"fold","sessionId":"s5"}`},
		{"next", `{"command":"next","sessionId":"s6"}`},
		{"hint", `{"command":"hint","sessionId":"s7"}`},
		{"log", `{"command":"log","sessionId":"s8"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.CrazyFourPokerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.CrazyFourPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		execRequest(t, ctrl.Exec, strings.NewReader("{invalid")).CodeIs(http.StatusBadRequest)
	})
}

// **Queens Up の省略は 0 として渡る。** 送らないことと「置かない」は同じ意味。
func TestCrazyFourPokerWebController_OmittedSideBetIsZero(t *testing.T) {
	m := new(usecase.MockCrazyFourPokerInteractor)
	m.On("PlaceBet", 50, 0).Return(`{}`)

	ctrl := controller.NewCrazyFourPokerWebController(func() uc.CrazyFourPokerInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.CrazyFourPokerWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","ante":50,"sessionId":"q1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertCalled(t, "PlaceBet", 50, 0)
}

// **必須パラメータの省略は 400。**
func TestCrazyFourPokerWebController_RequiredParams(t *testing.T) {
	for _, tt := range []struct{ name, body, want, notCalled string }{
		{"ante missing", `{"command":"bet","sessionId":"r1"}`, "ante is required", "PlaceBet"},
		{"multiplier missing", `{"command":"play","sessionId":"r2"}`, "multiplier is required", "Play"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := new(usecase.MockCrazyFourPokerInteractor)
			m.On("PlaceBet", 50, 0).Return(`{}`)
			m.On("Play", 1).Return(`{}`)
			ctrl := controller.NewCrazyFourPokerWebController(func() uc.CrazyFourPokerInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.CrazyFourPokerWebInput
			if err := json.Unmarshal([]byte(tt.body), &input); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
			if !strings.Contains(recorded.Body.String(), tt.want) {
				t.Errorf("expected %q, got: %s", tt.want, recorded.Body.String())
			}
			m.AssertNotCalled(t, tt.notCalled)
		})
	}
}

// **倍率の上限はコントローラで判定しない。** 手役次第なのでドメインへそのまま渡す。
func TestCrazyFourPokerWebController_MultiplierIsNotClampedHere(t *testing.T) {
	m := new(usecase.MockCrazyFourPokerInteractor)
	m.On("Play", 3).Return(`{}`)
	m.On("Play", 99).Return(`{"message":"needs aces"}`)

	ctrl := controller.NewCrazyFourPokerWebController(func() uc.CrazyFourPokerInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.CrazyFourPokerWebInput
	if err := json.Unmarshal([]byte(`{"command":"play","multiplier":99,"sessionId":"m1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	// 弾くのはドメインの仕事。コントローラは丸めずに渡すこと。
	m.AssertCalled(t, "Play", 99)
}
