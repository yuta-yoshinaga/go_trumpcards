//go:build test && (!js || !wasm || extra4)

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

func mustPutOutputJSON(msg string) string {
	out := &controller.PutWebOutput{
		Players:       []*controller.PutWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		MatchPoints:   []int{},
		WinnerIdx:     -1,
		HandWinnerIdx: -1,
		ResponderIdx:  -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPutOutputJSON: %v", err))
	}
	return string(b)
}

func TestPutWebController_Exec(t *testing.T) {
	mockOutput := `{"ok":true}`

	tiMock := new(usecase.MockPutInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultPutConfig()).Return(mockOutput)
	tiMock.On("Play", 2).Return(mockOutput)
	tiMock.On("Put").Return(mockOutput)
	tiMock.On("Respond", true).Return(mockOutput)
	tiMock.On("Respond", false).Return(mockOutput)
	tiMock.On("Next").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PutInteractorIF { return tiMock }
	ctrl := controller.NewPutWebController(factory)
	defer ctrl.Stop()

	exec := func(body string) *recorded {
		var input controller.PutWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit returns bye", func(t *testing.T) {
		rec := exec(`{"command":"q","sessionId":"s1"}`)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustPutOutputJSON("bye."))
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
	t.Run("put", func(t *testing.T) {
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

func TestPutWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.PutWebInput{}
		assert.Equal(t, domain.DefaultPutConfig(), input.ToConfig())
	})

	t.Run("explicit match target", func(t *testing.T) {
		mt := 30
		c := &controller.PutWebConfig{MatchTarget: &mt}
		assert.Equal(t, 30, c.ToConfig().MatchTarget)
	})

	t.Run("out-of-range match target clamps to default", func(t *testing.T) {
		mt := 9999
		c := &controller.PutWebConfig{MatchTarget: &mt}
		assert.Equal(t, domain.PutDefaultMatchTarget, c.ToConfig().MatchTarget)
	})

	t.Run("out-of-range difficulty clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.PutWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.PutCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})
}
