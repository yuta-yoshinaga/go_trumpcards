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

func mustGaigelOutputJSON(msg string) string {
	out := &controller.GaigelWebOutput{
		Players:         []*controller.GaigelWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		MarriageIndices: []int{},
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGaigelOutputJSON: %v", err))
	}
	return string(b)
}

func TestGaigelWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockGaigelInteractor)
	giMock.On("ResetWithConfig", domain.DefaultGaigelConfig()).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("DeclareMarriage", 1).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GaigelInteractorIF { return giMock }
	ctrl := controller.NewGaigelWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.GaigelWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustGaigelOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.GaigelWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play p with index", func(t *testing.T) {
		idx := 3
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("marriage m with index", func(t *testing.T) {
		idx := 1
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("marriage missing index", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "marriage", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("next n", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround nr", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.GaigelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestGaigelWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.GaigelWebInput{}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.GaigelCpuDifficultyNormal {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 101 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.GaigelCpuDifficultyHard)
		target := 51
		p := controller.GaigelWebInput{
			Config: &controller.GaigelWebConfig{
				CpuDifficulty: &cpu,
				TargetScore:   &target,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.GaigelCpuDifficultyHard {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 51 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
	})
}
