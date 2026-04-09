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

func mustPaiGowOutputJSON(msg string) string {
	out := &controller.PaiGowWebOutput{
		PlayerCards:    make([]*controller.WebOutputCard, 0),
		DealerCards:    make([]*controller.WebOutputCard, 0),
		PlayerHighHand: make([]*controller.WebOutputCard, 0),
		PlayerLowHand:  make([]*controller.WebOutputCard, 0),
		DealerHighHand: make([]*controller.WebOutputCard, 0),
		DealerLowHand:  make([]*controller.WebOutputCard, 0),
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPaiGowOutputJSON: %v", err))
	}
	return string(b)
}

func TestPaiGowWebController_Method(t *testing.T) {
	mockOutput := `{"playerCards":[],"dealerCards":[],"playerHighHand":[],"playerLowHand":[],"dealerHighHand":[],"dealerLowHand":[],"phase":0,"chips":0,"bet":0,"result":0,"highHandResult":0,"lowHandResult":0,"payout":0,"commission":0,"playerHighRank":0,"playerLowRank":0,"dealerHighRank":0,"dealerLowRank":0,"message":""}`
	expectedBody := mockOutput

	piMock := new(usecase.MockPaiGowInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("Bet", 100).Return(mockOutput)
	piMock.On("SetHands", 0, 1).Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PaiGowInteractorIF { return piMock }
	ctrl := controller.NewPaiGowWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustPaiGowOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet b", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("set", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"set","low0":0,"low1":1,"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("set s", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"s","low0":0,"low1":1,"sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log l", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.PaiGowWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s11"}`), &input)
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
