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

func mustRummy500OutputJSON(msg string) string {
	out := &controller.Rummy500WebOutput{
		Players:       []*controller.Rummy500WebOutputPlayer{},
		WinnerIdx:     -1,
		RoundEnderIdx: -1,
		DiscardPile:   []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustRummy500OutputJSON: %v", err))
	}
	return string(b)
}

func TestRummy500WebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	siMock := new(usecase.MockRummy500Interactor)
	siMock.On("ResetWithConfig", domain.DefaultRummy500Config()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard", -1).Return(mockOutput)
	siMock.On("DrawFromDiscard", 1).Return(mockOutput)
	siMock.On("Meld", []int{0, 1, 2}).Return(mockOutput)
	siMock.On("Layoff", 0, 0, 0).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.Rummy500InteractorIF { return siMock }
	ctrl := controller.NewRummy500WebController(factory)
	defer ctrl.Stop()

	t.Run("q quits", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustRummy500OutputJSON("bye."))
	})

	t.Run("r resets", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("ds draws stock", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"ds","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("dd default idx -1", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"dd","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("dd with discardIdx", func(t *testing.T) {
		idx := 1
		input := controller.Rummy500WebInput{
			BaseWebInput: controller.BaseWebInput{Command: "dd", SessionID: "r5-t-1"},
			DiscardIdx:   &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("m meld", func(t *testing.T) {
		input := controller.Rummy500WebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "r5-t-1"},
			CardIndices:  []int{0, 1, 2},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("lo layoff", func(t *testing.T) {
		owner := 0
		mIdx := 0
		cIdx := 0
		input := controller.Rummy500WebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "r5-t-1"},
			MeldOwner:    &owner,
			MeldIdx:      &mIdx,
			CardIndex:    &cIdx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("lo missing args", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"lo","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("d discard", func(t *testing.T) {
		idx := 3
		input := controller.Rummy500WebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "r5-t-1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("d missing card index", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("nr next round", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log fallthrough", func(t *testing.T) {
		var input controller.Rummy500WebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"r5-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestRummy500WebInput_ToConfig(t *testing.T) {
	t.Run("defaults when nil", func(t *testing.T) {
		var in controller.Rummy500WebInput
		cfg := in.ToConfig()
		if cfg != domain.DefaultRummy500Config() {
			t.Fatalf("expected default config got %+v", cfg)
		}
	})

	t.Run("bounded values", func(t *testing.T) {
		cd := int(domain.Rummy500CpuDifficultyHard)
		pl := 700
		in := controller.Rummy500WebInput{
			Config: &controller.Rummy500WebConfig{CpuDifficulty: &cd, PointLimit: &pl},
		}
		cfg := in.ToConfig()
		if cfg.CpuDifficulty != domain.Rummy500CpuDifficultyHard {
			t.Errorf("expected hard difficulty")
		}
		if cfg.PointLimit != 700 {
			t.Errorf("expected 700, got %d", cfg.PointLimit)
		}
	})

	t.Run("clamps out-of-range", func(t *testing.T) {
		cd := 99
		pl := -10
		in := controller.Rummy500WebInput{
			Config: &controller.Rummy500WebConfig{CpuDifficulty: &cd, PointLimit: &pl},
		}
		cfg := in.ToConfig()
		if cfg.CpuDifficulty < domain.Rummy500CpuDifficultyEasy || cfg.CpuDifficulty > domain.Rummy500CpuDifficultyHard {
			t.Errorf("difficulty not clamped: %d", cfg.CpuDifficulty)
		}
		if cfg.PointLimit < 1 {
			t.Errorf("point limit not clamped: %d", cfg.PointLimit)
		}
	})
}
