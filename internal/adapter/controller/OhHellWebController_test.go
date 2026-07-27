//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustOhHellOutputJSON(msg string) string {
	out := &controller.OhHellWebOutput{
		Players:       []*controller.OhHellWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		TrumpSuit:     -1,
		RestrictedBid: -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOhHellOutputJSON: %v", err))
	}
	return string(b)
}

func TestOhHellWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"totalRounds":0,"handSize":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"dealerIdx":0,"trumpCard":null,"trumpSuit":-1,"restrictedBid":-1,"currentTrick":[],"gameEndFlag":false,"winnerIdx":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"maxHandSize":0,"scoringVariant":0,"roundDirection":0}}`
	expectedBody := mockOutput

	oiMock := new(usecase.MockOhHellInteractor)
	oiMock.On("ResetWithConfig", domain.DefaultOhHellConfig()).Return(mockOutput)
	oiMock.On("Bid", 3).Return(mockOutput)
	oiMock.On("Play", 3).Return(mockOutput)
	oiMock.On("NextTrick").Return(mockOutput)
	oiMock.On("NextRound").Return(mockOutput)
	oiMock.On("Hint").Return(mockOutput)
	oiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OhHellInteractorIF { return oiMock }
	ctrl := controller.NewOhHellWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOhHellOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec b", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","bid":3,"sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bid missing param", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("success Exec p", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","cardIndex":3,"sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play missing param", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("success Exec n", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.OhHellWebInput
		_ = json.Unmarshal([]byte(`{"command":"xxx","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestOhHellWebConfig_ToConfig(t *testing.T) {
	t.Run("all nil uses defaults", func(t *testing.T) {
		cfg := (&controller.OhHellWebConfig{}).ToConfig()
		expected := domain.DefaultOhHellConfig()
		if cfg != expected {
			t.Errorf("expected %+v, got %+v", expected, cfg)
		}
	})

	t.Run("values are bounded", func(t *testing.T) {
		v99 := 99
		cfg := (&controller.OhHellWebConfig{CpuDifficulty: &v99}).ToConfig()
		if int(cfg.CpuDifficulty) > 2 {
			t.Errorf("expected bounded difficulty, got %d", cfg.CpuDifficulty)
		}
	})
}
