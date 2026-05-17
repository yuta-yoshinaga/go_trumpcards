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

func mustFourCardPokerOutputJSON(msg string) string {
	out := &controller.FourCardPokerWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		PlayerBest:    make([]*controller.WebOutputCard, 0),
		DealerBest:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFourCardPokerOutputJSON: %v", err))
	}
	return string(b)
}

func TestFourCardPokerWebController_Method(t *testing.T) {
	mockOutput := mustFourCardPokerOutputJSON("")
	expectedBody := mockOutput

	tiMock := new(usecase.MockFourCardPokerInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Bet", 100, 0).Return(mockOutput)
	tiMock.On("Bet", 100, 50).Return(mockOutput)
	tiMock.On("Play", 1).Return(mockOutput)
	tiMock.On("Play", 2).Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("Fold").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.FourCardPokerInteractorIF { return tiMock }
	ctrl := controller.NewFourCardPokerWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustFourCardPokerOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet ante only", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet b short", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet with acesUp", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"acesUpBet":50,"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play default 1x", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play 2x", func(t *testing.T) {
		mul := 2
		input := controller.FourCardPokerWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "p", SessionID: "s9"},
			PlayMultiplier: &mul,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play 3x", func(t *testing.T) {
		mul := 3
		input := controller.FourCardPokerWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "p", SessionID: "s10"},
			PlayMultiplier: &mul,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold f", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s13"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.FourCardPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s14"}`), &input)
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
