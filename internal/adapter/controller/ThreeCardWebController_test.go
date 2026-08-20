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

func mustThreeCardOutputJSON(msg string) string {
	out := &controller.ThreeCardWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		DealerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustThreeCardOutputJSON: %v", err))
	}
	return string(b)
}

func threeCardIntPtr(v int) *int { return &v }

func TestThreeCardWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"dealerHand":[],"phase":0,"chips":0,"anteBet":0,"pairPlusBet":0,"playBet":0,"result":0,"antePayout":0,"playPayout":0,"anteBonusPayout":0,"pairPlusPayout":0,"totalPayout":0,"dealerQualified":false,"playerHandRank":0,"dealerHandRank":0,"message":""}`
	expectedBody := mockOutput

	tiMock := new(usecase.MockThreeCardInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Bet", 100, 0).Return(mockOutput)
	tiMock.On("Bet", 100, 50).Return(mockOutput)
	tiMock.On("Rebet").Return(mockOutput)
	tiMock.On("Play").Return(mockOutput)
	tiMock.On("Fold").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ThreeCardInteractorIF { return tiMock }
	ctrl := controller.NewThreeCardWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustThreeCardOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// #5513: 直前と同じ額で賭け直す。額はサーバが覚えているので送らない。
	t.Run("rebet", func(t *testing.T) {
		for _, cmd := range []string{"rebet", "rb"} {
			var input controller.ThreeCardWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"s-rebet"}`), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(expectedBody)
		}
	})

	t.Run("bet b", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet with pairPlus", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"pairPlusBet":50,"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play p", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold f", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log l", func(t *testing.T) {
		var input controller.ThreeCardWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s13"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.ThreeCardWebInput
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

	_ = threeCardIntPtr(0) // suppress unused lint
}
