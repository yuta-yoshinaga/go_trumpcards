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

func mustSchnapsenOutputJSON(msg string) string {
	out := &controller.SchnapsenWebOutput{
		Players:       []*controller.SchnapsenWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		MarriagePlays: []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSchnapsenOutputJSON: %v", err))
	}
	return string(b)
}

func TestSchnapsenWebController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockSchnapsenInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSchnapsenConfig()).Return(mockOutput)
	siMock.On("Play", 2).Return(mockOutput)
	siMock.On("DeclareMarriage", 3).Return(mockOutput)
	siMock.On("NextTrick").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SchnapsenInteractorIF { return siMock }
	ctrl := controller.NewSchnapsenWebController(factory)
	defer ctrl.Stop()

	t.Run("quit returns bye", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustSchnapsenOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play with cardIndex", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1","cardIndex":2}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play without cardIndex", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "cardIndex is required")
	})

	t.Run("marriage with cardIndex", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","cardIndex":3}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("marriage without cardIndex", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "cardIndex is required")
	})

	t.Run("next", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.SchnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
}

func TestSchnapsenWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.SchnapsenWebInput{}
		assert.Equal(t, domain.DefaultSchnapsenConfig(), input.ToConfig())
	})

	t.Run("explicit normal difficulty", func(t *testing.T) {
		diff := int(domain.SchnapsenCpuDifficultyNormal)
		c := &controller.SchnapsenWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.SchnapsenCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})

	t.Run("out-of-range clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.SchnapsenWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.SchnapsenCpuDifficultyNormal, c.ToConfig().CpuDifficulty)
	})
}
