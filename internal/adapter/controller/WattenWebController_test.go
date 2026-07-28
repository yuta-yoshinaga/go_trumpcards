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

func mustWattenOutputJSON(msg string) string {
	out := &controller.WattenWebOutput{
		Players:        []*controller.WattenWebOutputPlayer{},
		CurrentTrick:   []*controller.WebOutputTrickCard{},
		WinnerTeam:     -1,
		RaiserTeam:     -1,
		ResponderIdx:   -1,
		DealWinnerTeam: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWattenOutputJSON: %v", err))
	}
	return string(b)
}

func TestWattenWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	wiMock := new(usecase.MockWattenInteractor)
	wiMock.On("ResetWithConfig", domain.DefaultWattenConfig()).Return(mockOutput)
	wiMock.On("Declare", 10, 3).Return(mockOutput)
	wiMock.On("Play", 3).Return(mockOutput)
	wiMock.On("Raise").Return(mockOutput)
	wiMock.On("Respond", true).Return(mockOutput)
	wiMock.On("NextRound").Return(mockOutput)
	wiMock.On("Hint").Return(mockOutput)
	wiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.WattenInteractorIF { return wiMock }
	ctrl := controller.NewWattenWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.WattenWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustWattenOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.WattenWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("declare", func(t *testing.T) {
		rank, suit := 10, 3
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			Rank:         &rank,
			Suit:         &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("declare missing", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "declare", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("play", func(t *testing.T) {
		idx := 3
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("raise", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "rz", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("respond", func(t *testing.T) {
		hold := true
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "respond", SessionID: "s1"},
			Hold:         &hold,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("respond missing", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "resp", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("nextround", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.WattenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestWattenWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.WattenWebInput{}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.WattenCpuDifficultyNormal {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 15 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
		if cfg.MaxRaises != 5 {
			t.Errorf("maxRaises: got %d", cfg.MaxRaises)
		}
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.WattenCpuDifficultyHard)
		target := 21
		raises := 3
		p := controller.WattenWebInput{
			Config: &controller.WattenWebConfig{
				CpuDifficulty: &cpu,
				TargetScore:   &target,
				MaxRaises:     &raises,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.WattenCpuDifficultyHard {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 21 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
		if cfg.MaxRaises != 3 {
			t.Errorf("maxRaises: got %d", cfg.MaxRaises)
		}
	})
}
