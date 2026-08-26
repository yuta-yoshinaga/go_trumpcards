//go:build test && (!js || !wasm || classic)

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustBrusquembilleOutputJSON(msg string) string {
	out := &controller.BrusquembilleWebOutput{
		Players:       []*controller.BrusquembilleWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBrusquembilleOutputJSON: %v", err))
	}
	return string(b)
}

func TestBrusquembilleWebController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"trickNumber":0,"currentPlayerIdx":0,"currentTrick":[],"trumpSuit":0,"dealerIdx":0,"leadPlayerIdx":0,"stockRemaining":0,"gameEndFlag":false,"winnerIdx":-1,"message":"","config":{"cpuDifficulty":0}}`

	biMock := new(usecase.MockBrusquembilleInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBrusquembilleConfig()).Return(mockOutput)
	biMock.On("Play", 2).Return(mockOutput)
	biMock.On("NextTrick").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BrusquembilleInteractorIF { return biMock }
	ctrl := controller.NewBrusquembilleWebController(factory)
	defer ctrl.Stop()

	t.Run("quit returns bye", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustBrusquembilleOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play with cardIndex", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1","cardIndex":2}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play without cardIndex", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(rec.Body.String())
		assert.Contains(t, body, "cardIndex is required")
	})

	t.Run("next", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.BrusquembilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
}

func TestBrusquembilleWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.BrusquembilleWebInput{}
		assert.Equal(t, domain.DefaultBrusquembilleConfig(), input.ToConfig())
	})

	t.Run("explicit normal difficulty", func(t *testing.T) {
		diff := int(domain.BrusquembilleCpuDifficultyNormal)
		c := &controller.BrusquembilleWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.BrusquembilleCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})

	t.Run("out-of-range clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.BrusquembilleWebConfig{CpuDifficulty: &diff}
		// v1 only supports Normal; out-of-range falls back to the default.
		assert.Equal(t, domain.BrusquembilleCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})
}
