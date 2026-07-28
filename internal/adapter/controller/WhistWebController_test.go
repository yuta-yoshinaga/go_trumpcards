//go:build test

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

func mustWhistOutputJSON(msg string) string {
	out := &controller.WhistWebOutput{
		Players:       []*controller.WhistWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWhistOutputJSON: %v", err))
	}
	return string(b)
}

func TestWhistWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"currentTrick":[],"trumpSuit":0,"dealerIdx":0,"teamScores":[0,0],"gameEndFlag":false,"winnerTeam":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	wiMock := new(usecase.MockWhistInteractor)
	wiMock.On("ResetWithConfig", domain.DefaultWhistConfig()).Return(mockOutput)
	wiMock.On("Play", 3).Return(mockOutput)
	wiMock.On("NextTrick").Return(mockOutput)
	wiMock.On("NextRound").Return(mockOutput)
	wiMock.On("Hint").Return(mockOutput)
	wiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.WhistInteractorIF { return wiMock }
	ctrl := controller.NewWhistWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustWhistOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1","cardIndex":3}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("play without cardIndex", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(recorded.Body.String())
		assert.Contains(t, body, "cardIndex is required")
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.WhistWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
}

func TestWhistWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.WhistWebInput{}
		cfg := input.ToConfig()
		expected := domain.DefaultWhistConfig()
		assert.Equal(t, expected, cfg)
	})

	t.Run("custom config", func(t *testing.T) {
		diff := 2
		limit := 10
		c := &controller.WhistWebConfig{CpuDifficulty: &diff, PointLimit: &limit}
		cfg := c.ToConfig()
		assert.Equal(t, domain.WhistCpuDifficultyHard, cfg.CpuDifficulty)
		assert.Equal(t, 10, cfg.PointLimit)
	})
}
