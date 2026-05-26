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

func mustRussianPokerOutputJSON(msg string) string {
	out := &controller.RussianPokerWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustRussianPokerOutputJSON: %v", err))
	}
	return string(b)
}

func TestRussianPokerWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"phase":0,"chips":0,"anteBet":0,"exchangeCount":0,"exchangeFee":0,"bought6th":false,"buy6thFee":0,"forceExchanged":false,"forceExchangeFee":0,"playBet":0,"result":0,"antePayout":0,"playPayout":0,"totalPayout":0,"dealerQualified":false,"playerHandRank":0,"dealerHandRank":0,"message":""}`
	expectedBody := mockOutput

	riMock := new(usecase.MockRussianPokerInteractor)
	riMock.On("Reset").Return(mockOutput)
	riMock.On("Bet", 100).Return(mockOutput)
	riMock.On("Exchange", []int(nil)).Return(mockOutput)
	riMock.On("Exchange", []int{0, 2}).Return(mockOutput)
	riMock.On("Buy6th").Return(mockOutput)
	riMock.On("Select", -1).Return(mockOutput)
	riMock.On("Select", 3).Return(mockOutput)
	riMock.On("Play").Return(mockOutput)
	riMock.On("Fold").Return(mockOutput)
	riMock.On("ForceExchange").Return(mockOutput)
	riMock.On("Decline").Return(mockOutput)
	riMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.RussianPokerInteractorIF { return riMock }
	ctrl := controller.NewRussianPokerWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustRussianPokerOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("exchange", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"exchange","indices":[0,2],"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("exchange e empty", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"e","sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("buy6th", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"buy6th","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("buy6th 6", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"6","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("select with index", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"select","discardIndex":3,"sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("select without index", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"sel","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s13"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("force", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"force","sessionId":"s14"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("force fe", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"fe","sessionId":"s15"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("decline", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"decline","sessionId":"s16"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("decline d", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s17"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s18"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.RussianPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s19"}`), &input)
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
