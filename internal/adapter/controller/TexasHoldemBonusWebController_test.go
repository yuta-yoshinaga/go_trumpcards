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

func mustTexasHoldemBonusOutputJSON(msg string) string {
	out := &controller.TexasHoldemBonusWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		Community:     make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTexasHoldemBonusOutputJSON: %v", err))
	}
	return string(b)
}

func TestTexasHoldemBonusWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"community":[],"phase":0,"chips":0,"anteBet":0,"bonusBet":0,"flopBet":0,"turnBet":0,"riverBet":0,"totalPlayBet":0,"result":0,"antePayout":0,"playPayout":0,"bonusPayout":0,"totalPayout":0,"playerHandRank":0,"dealerHandRank":0,"message":""}`

	tiMock := new(usecase.MockTexasHoldemBonusInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Bet", 100, 0).Return(mockOutput)
	tiMock.On("Bet", 100, 10).Return(mockOutput)
	tiMock.On("Play").Return(mockOutput)
	tiMock.On("Fold").Return(mockOutput)
	tiMock.On("Check").Return(mockOutput)
	tiMock.On("Raise").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TexasHoldemBonusInteractorIF { return tiMock }
	ctrl := controller.NewTexasHoldemBonusWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"thb1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustTexasHoldemBonusOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"thb2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"thb3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("bet with bonus", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"bonusBet":10,"sessionId":"thb4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("play", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"thb5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"thb6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("check", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"check","sessionId":"thb7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("raise long", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"raise","sessionId":"thb8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("raise short", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"ra","sessionId":"thb9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"thb10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.TexasHoldemBonusWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"thb11"}`), &input)
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
