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

func mustCasinoHoldemOutputJSON(msg string) string {
	out := &controller.CasinoHoldemWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		Community:     make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCasinoHoldemOutputJSON: %v", err))
	}
	return string(b)
}

func TestCasinoHoldemWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"community":[],"phase":0,"chips":0,"anteBet":0,"bonusBet":0,"callBet":0,"result":0,"dealerQualify":false,"antePayout":0,"callPayout":0,"bonusPayout":0,"totalPayout":0,"playerHandRank":0,"dealerHandRank":0,"message":""}`

	ciMock := new(usecase.MockCasinoHoldemInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Bet", 100, 0).Return(mockOutput)
	ciMock.On("Bet", 100, 10).Return(mockOutput)
	ciMock.On("Call").Return(mockOutput)
	ciMock.On("Fold").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CasinoHoldemInteractorIF { return ciMock }
	ctrl := controller.NewCasinoHoldemWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"ch1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustCasinoHoldemOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"ch2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"ch3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("bet with bonus", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"bonusBet":10,"sessionId":"ch4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("call long", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"call","sessionId":"ch5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("call short", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"c","sessionId":"ch6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"ch7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"ch8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.CasinoHoldemWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"ch9"}`), &input)
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
