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

func mustHighCardFlushOutputJSON(msg string) string {
	out := &controller.HighCardFlushWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHighCardFlushOutputJSON: %v", err))
	}
	return string(b)
}

func TestHighCardFlushWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"phase":0,"chips":0,"anteBet":0,"flushBonusBet":0,"straightFlushBet":0,"raiseBet":0,"result":0,"antePayout":0,"raisePayout":0,"flushBonusPayout":0,"straightFlushPayout":0,"totalPayout":0,"dealerQualified":false,"playerFlushLen":0,"dealerFlushLen":0,"playerStraightFlushLen":0,"maxRaiseMultiplier":0,"message":""}`
	expectedBody := mockOutput

	hiMock := new(usecase.MockHighCardFlushInteractor)
	hiMock.On("Reset").Return(mockOutput)
	hiMock.On("Bet", 100, 0, 0).Return(mockOutput)
	hiMock.On("Bet", 100, 50, 20).Return(mockOutput)
	hiMock.On("Raise", 1).Return(mockOutput)
	hiMock.On("Raise", 2).Return(mockOutput)
	hiMock.On("Raise", 3).Return(mockOutput)
	hiMock.On("Fold").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.HighCardFlushInteractorIF { return hiMock }
	ctrl := controller.NewHighCardFlushWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustHighCardFlushOutputJSON("bye."))
	})
	t.Run("quit", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})
	t.Run("reset r", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("bet ante only", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})
	t.Run("bet b shortform", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("bet with side bets", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"flushBonusBet":50,"straightFlushBet":20,"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("raise multiplier 1", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"raise","multiplier":1,"sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("raise ra shortform", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"ra","multiplier":2,"sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("fold", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("fold f", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("unknown command", func(t *testing.T) {
		var input controller.HighCardFlushWebInput
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
