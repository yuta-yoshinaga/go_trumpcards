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

func mustPanOutputJSON(msg string) string {
	out := &controller.PanWebOutput{
		Players:        []*controller.PanWebOutputPlayer{},
		WinnerIdx:      -1,
		PanDeclarerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPanOutputJSON: %v", err))
	}
	return string(b)
}

func TestPanWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	siMock := new(usecase.MockPanInteractor)
	siMock.On("ResetWithConfig", domain.DefaultPanConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Meld", []int{0, 1, 2}).Return(mockOutput)
	siMock.On("Layoff", 1, 0, 2).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PanInteractorIF { return siMock }
	ctrl := controller.NewPanWebController(factory)
	defer ctrl.Stop()

	t.Run("q quits", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustPanOutputJSON("bye."))
	})

	t.Run("r resets", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("ds draws stock", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"ds","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("dd draws discard", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"dd","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("m meld", func(t *testing.T) {
		input := controller.PanWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "pan-t-1"},
			CardIndices:  []int{0, 1, 2},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("lo layoff", func(t *testing.T) {
		owner := 1
		mIdx := 0
		cIdx := 2
		input := controller.PanWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "pan-t-1"},
			MeldOwner:    &owner,
			MeldIdx:      &mIdx,
			CardIndex:    &cIdx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("lo missing args", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"lo","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("d discard", func(t *testing.T) {
		idx := 3
		input := controller.PanWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "pan-t-1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("d missing card index", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("nr next round", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log fallthrough", func(t *testing.T) {
		var input controller.PanWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"pan-t-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestPanWebInput_ToConfig(t *testing.T) {
	t.Run("defaults when nil", func(t *testing.T) {
		var in controller.PanWebInput
		if in.ToConfig() != domain.DefaultPanConfig() {
			t.Fatalf("expected default config")
		}
	})

	t.Run("bounded values", func(t *testing.T) {
		pc := 6
		cd := int(domain.PanCpuDifficultyHard)
		tr := 5
		in := controller.PanWebInput{
			Config: &controller.PanWebConfig{PlayerCount: &pc, CpuDifficulty: &cd, TargetRounds: &tr},
		}
		cfg := in.ToConfig()
		if cfg.PlayerCount != 6 || cfg.CpuDifficulty != domain.PanCpuDifficultyHard || cfg.TargetRounds != 5 {
			t.Errorf("unexpected config %+v", cfg)
		}
	})

	t.Run("clamps out-of-range", func(t *testing.T) {
		pc := 99
		cd := 99
		tr := -3
		in := controller.PanWebInput{
			Config: &controller.PanWebConfig{PlayerCount: &pc, CpuDifficulty: &cd, TargetRounds: &tr},
		}
		cfg := in.ToConfig()
		if cfg.PlayerCount < domain.PanPlayerCountMin || cfg.PlayerCount > domain.PanPlayerCountMax {
			t.Errorf("player count not clamped: %d", cfg.PlayerCount)
		}
		if cfg.CpuDifficulty < domain.PanCpuDifficultyEasy || cfg.CpuDifficulty > domain.PanCpuDifficultyHard {
			t.Errorf("difficulty not clamped: %d", cfg.CpuDifficulty)
		}
		if cfg.TargetRounds < 1 {
			t.Errorf("target rounds not clamped: %d", cfg.TargetRounds)
		}
	})
}
