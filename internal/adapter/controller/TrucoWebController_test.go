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

func mustTrucoOutputJSON(msg string) string {
	out := &controller.TrucoWebOutput{
		Players:       []*controller.TrucoWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		MatchPoints:   []int{},
		WinnerIdx:     -1,
		HandWinnerIdx: -1,
		ResponderIdx:  -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTrucoOutputJSON: %v", err))
	}
	return string(b)
}

func TestTrucoWebController_Exec(t *testing.T) {
	mockOutput := `{"ok":true}`

	tiMock := new(usecase.MockTrucoInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTrucoConfig()).Return(mockOutput)
	tiMock.On("Play", 2).Return(mockOutput)
	tiMock.On("Truco").Return(mockOutput)
	tiMock.On("Respond", true).Return(mockOutput)
	tiMock.On("Respond", false).Return(mockOutput)
	tiMock.On("Next").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TrucoInteractorIF { return tiMock }
	ctrl := controller.NewTrucoWebController(factory)
	defer ctrl.Stop()

	exec := func(body string) *recorded {
		var input controller.TrucoWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit returns bye", func(t *testing.T) {
		rec := exec(`{"command":"q","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustTrucoOutputJSON("bye."))
	})
	t.Run("reset", func(t *testing.T) {
		rec := exec(`{"command":"r","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("play with cardIndex", func(t *testing.T) {
		rec := exec(`{"command":"p","sessionId":"s1","cardIndex":2}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("play without cardIndex", func(t *testing.T) {
		rec := exec(`{"command":"p","sessionId":"s1"}`)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "cardIndex is required")
	})
	t.Run("truco", func(t *testing.T) {
		rec := exec(`{"command":"t","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("accept", func(t *testing.T) {
		rec := exec(`{"command":"a","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("decline", func(t *testing.T) {
		rec := exec(`{"command":"d","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("next", func(t *testing.T) {
		rec := exec(`{"command":"n","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("hint", func(t *testing.T) {
		rec := exec(`{"command":"hint","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
	t.Run("log", func(t *testing.T) {
		rec := exec(`{"command":"log","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
}

func TestTrucoWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.TrucoWebInput{}
		assert.Equal(t, domain.DefaultTrucoConfig(), input.ToConfig())
	})

	t.Run("explicit match target", func(t *testing.T) {
		mt := 30
		c := &controller.TrucoWebConfig{MatchTarget: &mt}
		assert.Equal(t, 30, c.ToConfig().MatchTarget)
	})

	t.Run("out-of-range match target clamps to default", func(t *testing.T) {
		mt := 9999
		c := &controller.TrucoWebConfig{MatchTarget: &mt}
		assert.Equal(t, domain.TrucoDefaultMatchTarget, c.ToConfig().MatchTarget)
	})

	t.Run("out-of-range difficulty clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.TrucoWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.TrucoCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})
}
