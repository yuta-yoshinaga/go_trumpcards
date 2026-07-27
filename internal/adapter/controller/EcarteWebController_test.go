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

func mustEcarteOutputJSON(msg string) string {
	out := &controller.EcarteWebOutput{
		Players:       []*controller.EcarteWebOutputPlayer{},
		DealPoints:    []int{},
		MatchScore:    []int{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustEcarteOutputJSON: %v", err))
	}
	return string(b)
}

func TestEcarteWebController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	eiMock := new(usecase.MockEcarteInteractor)
	eiMock.On("ResetWithConfig", domain.DefaultEcarteConfig()).Return(mockOutput)
	eiMock.On("Propose").Return(mockOutput)
	eiMock.On("Stand").Return(mockOutput)
	eiMock.On("Respond", true).Return(mockOutput)
	eiMock.On("Respond", false).Return(mockOutput)
	eiMock.On("Discard", []int{0, 2}).Return(mockOutput)
	eiMock.On("Discard", []int{}).Return(mockOutput)
	eiMock.On("Play", 2).Return(mockOutput)
	eiMock.On("NextRound").Return(mockOutput)
	eiMock.On("Hint").Return(mockOutput)
	eiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.EcarteInteractorIF { return eiMock }
	ctrl := controller.NewEcarteWebController(factory)
	defer ctrl.Stop()

	t.Run("quit returns bye", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustEcarteOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("propose", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"propose","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("stand", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"stand","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("respond accept", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"respond","sessionId":"s1","accept":true}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("respond refuse", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"respond","sessionId":"s1","accept":false}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("respond without accept", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"respond","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "accept is required")
	})

	t.Run("discard with indices", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"discard","sessionId":"s1","discardIndices":[0,2]}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("discard without indices", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"discard","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play with cardIndex", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1","cardIndex":2}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play without cardIndex", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		assert.Contains(t, strings.TrimSpace(rec.Body.String()), "cardIndex is required")
	})

	t.Run("next", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.EcarteWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})
}

func TestEcarteWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		input := controller.EcarteWebInput{}
		assert.Equal(t, domain.DefaultEcarteConfig(), input.ToConfig())
	})

	t.Run("explicit hard difficulty and target", func(t *testing.T) {
		diff := int(domain.EcarteCpuDifficultyHard)
		target := 10
		c := &controller.EcarteWebConfig{CpuDifficulty: &diff, TargetScore: &target}
		got := c.ToConfig()
		assert.Equal(t, domain.EcarteCpuDifficultyHard, got.CpuDifficulty)
		assert.Equal(t, 10, got.TargetScore)
	})

	t.Run("out-of-range difficulty clamps to default", func(t *testing.T) {
		diff := 99
		c := &controller.EcarteWebConfig{CpuDifficulty: &diff}
		assert.Equal(t, domain.DefaultEcarteConfig().CpuDifficulty, c.ToConfig().CpuDifficulty)
	})
}
