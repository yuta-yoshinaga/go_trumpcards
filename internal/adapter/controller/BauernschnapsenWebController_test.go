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

func mustBauernschnapsenOutputJSON(msg string) string {
	out := &controller.BauernschnapsenWebOutput{
		Players:          []*controller.BauernschnapsenWebOutputPlayer{},
		CurrentTrick:     []*controller.WebOutputTrickCard{},
		ValidPlayIndices: []int{},
		MarriageIndices:  []int{},
		WinnerTeam:       -1,
		WebOutputBase:    controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBauernschnapsenOutputJSON: %v", err))
	}
	return string(b)
}

func TestBauernschnapsenWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockBauernschnapsenInteractor)
	giMock.On("ResetWithConfig", domain.DefaultBauernschnapsenConfig()).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("DeclareMarriage", 1).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BauernschnapsenInteractorIF { return giMock }
	ctrl := controller.NewBauernschnapsenWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.BauernschnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustBauernschnapsenOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BauernschnapsenWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play p with index", func(t *testing.T) {
		idx := 3
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("marriage m with index", func(t *testing.T) {
		idx := 1
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("marriage missing index", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "marriage", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("next n", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround nr", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.BauernschnapsenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestBauernschnapsenWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.BauernschnapsenWebInput{}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.BauernschnapsenCpuDifficultyNormal {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 24 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.BauernschnapsenCpuDifficultyHard)
		target := 51
		p := controller.BauernschnapsenWebInput{
			Config: &controller.BauernschnapsenWebConfig{
				CpuDifficulty: &cpu,
				TargetScore:   &target,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.BauernschnapsenCpuDifficultyHard {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 51 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
	})
}
