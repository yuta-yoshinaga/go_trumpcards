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

func mustBeziqueOutputJSON(msg string) string {
	out := &controller.BeziqueWebOutput{
		Players:        []*controller.BeziqueWebOutputPlayer{},
		DealPoints:     []int{},
		DealMeldPoints: []int{},
		MatchScore:     []int{},
		CurrentTrick:   []*controller.WebOutputTrickCard{},
		AvailableMelds: []*controller.BeziqueWebOutputMeld{},
		WinnerIdx:      -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBeziqueOutputJSON: %v", err))
	}
	return string(b)
}

func TestBeziqueWebController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	biMock := new(usecase.MockBeziqueInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBeziqueConfig()).Return(mockOutput)
	biMock.On("Play", 2).Return(mockOutput)
	biMock.On("DeclareMeld", 1).Return(mockOutput)
	biMock.On("SkipMeld").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BeziqueInteractorIF { return biMock }
	ctrl := controller.NewBeziqueWebController(factory)
	defer ctrl.Stop()

	t.Run("quit returns bye", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustBeziqueOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play with cardIndex", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1","cardIndex":2}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play without cardIndex", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "cardIndex is required")
	})

	t.Run("meld with meldIndex", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","meldIndex":1}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("meld without meldIndex", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "meldIndex is required")
	})

	t.Run("skip", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"s","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("next", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.BeziqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
}

func TestBeziqueWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.BeziqueWebInput{}
		assert.Equal(t, domain.DefaultBeziqueConfig(), input.ToConfig())
	})

	t.Run("explicit hard difficulty and target", func(t *testing.T) {
		diff := int(domain.BeziqueCpuDifficultyHard)
		target := 500
		c := &controller.BeziqueWebConfig{CpuDifficulty: &diff, TargetScore: &target}
		got := c.ToConfig()
		assert.Equal(t, domain.BeziqueCpuDifficultyHard, got.CpuDifficulty)
		assert.Equal(t, 500, got.TargetScore)
	})

	t.Run("out-of-range difficulty clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.BeziqueWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.DefaultBeziqueConfig().CpuDifficulty, c.ToConfig().CpuDifficulty)
	})
}
