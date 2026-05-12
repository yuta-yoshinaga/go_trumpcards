//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustUltimateTexasHoldemOutputJSON(msg string) string {
	out := &controller.UltimateTexasHoldemWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		Community:     make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustUltimateTexasHoldemOutputJSON: %v", err))
	}
	return string(b)
}

func TestUltimateTexasHoldemWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"community":[],"phase":0,"chips":0,"anteBet":0,"blindBet":0,"tripsBet":0,"playBet":0,"folded":false,"result":0,"dealerQualified":false,"antePayout":0,"blindPayout":0,"playPayout":0,"tripsPayout":0,"totalPayout":0,"playerHandRank":0,"dealerHandRank":0,"message":""}`

	tiMock := new(usecase.MockUltimateTexasHoldemInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Bet", 100, 0).Return(mockOutput)
	tiMock.On("Bet", 100, 10).Return(mockOutput)
	tiMock.On("Play", 4).Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("Play", 2).Return(mockOutput)
	tiMock.On("Play", 1).Return(mockOutput)
	tiMock.On("Check").Return(mockOutput)
	tiMock.On("Fold").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.UltimateTexasHoldemInteractorIF { return tiMock }
	ctrl := controller.NewUltimateTexasHoldemWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"uth1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustUltimateTexasHoldemOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"uth2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"uth3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("bet with trips", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"tripsBet":10,"sessionId":"uth4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("play preflop 4x", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","multiplier":4,"sessionId":"uth5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("play preflop 3x short", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","multiplier":3,"sessionId":"uth6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("play flop 2x", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","multiplier":2,"sessionId":"uth7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("play river 1x", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","multiplier":1,"sessionId":"uth8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("check", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"check","sessionId":"uth9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("fold long", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"uth10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("fold short", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"uth11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"uth12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.UltimateTexasHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"uth13"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(recorded.Body.String())
		if !strings.Contains(body, "Unsupported command") {
			t.Errorf("expected Unsupported command in body, got: %s", body)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
	})
}
