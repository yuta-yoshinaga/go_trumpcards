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

func mustOasisPokerOutputJSON(msg string) string {
	out := &controller.OasisPokerWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOasisPokerOutputJSON: %v", err))
	}
	return string(b)
}

func TestOasisPokerWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"phase":0,"chips":0,"anteBet":0,"jackpotBet":0,"exchangeCount":0,"exchangeFee":0,"playBet":0,"result":0,"antePayout":0,"playPayout":0,"jackpotPayout":0,"totalPayout":0,"dealerQualified":false,"playerHandRank":0,"dealerHandRank":0,"message":""}`
	expectedBody := mockOutput

	oiMock := new(usecase.MockOasisPokerInteractor)
	oiMock.On("Reset").Return(mockOutput)
	oiMock.On("Bet", 100, 0).Return(mockOutput)
	oiMock.On("Bet", 100, 10).Return(mockOutput)
	oiMock.On("Exchange", []int(nil)).Return(mockOutput)
	oiMock.On("Exchange", []int{0, 2}).Return(mockOutput)
	oiMock.On("Stand").Return(mockOutput)
	oiMock.On("Play").Return(mockOutput)
	oiMock.On("Fold").Return(mockOutput)
	oiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OasisPokerInteractorIF { return oiMock }
	ctrl := controller.NewOasisPokerWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustOasisPokerOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet with jackpot", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"jackpotBet":10,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("exchange", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"exchange","indices":[0,2],"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("exchange e empty", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"e","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("stand", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"stand","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.OasisPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s13"}`), &input)
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
