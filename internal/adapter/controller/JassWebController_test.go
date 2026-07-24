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

func mustJassOutputJSON(msg string) string {
	out := &controller.JassWebOutput{
		Players:         []*controller.JassWebOutputPlayer{},
		CurrentTrick:    []*controller.JassWebOutputTrickCard{},
		LastTrick:       []*controller.JassWebOutputTrickCard{},
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustJassOutputJSON: %v", err))
	}
	return string(b)
}

func TestJassWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	jiMock := new(usecase.MockJassInteractor)
	jiMock.On("ResetWithConfig", domain.DefaultJassConfig()).Return(mockOutput)
	jiMock.On("ChooseTrump", 2).Return(mockOutput)
	jiMock.On("Schieben").Return(mockOutput)
	jiMock.On("Play", 3).Return(mockOutput)
	jiMock.On("NextTrick").Return(mockOutput)
	jiMock.On("NextRound").Return(mockOutput)
	jiMock.On("Hint").Return(mockOutput)
	jiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.JassInteractorIF { return jiMock }
	ctrl := controller.NewJassWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.JassWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustJassOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.JassWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("calltrump c", func(t *testing.T) {
		suit := 2
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "s1"},
			Suit:         &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("calltrump missing suit", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "calltrump", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("schieben sc", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "sc", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play p with index", func(t *testing.T) {
		idx := 3
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("next n", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround nr", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.JassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestJassWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.JassWebInput{}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.JassCpuDifficultyNormal {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 1000 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
		if cfg.LastTrickBonus != 5 {
			t.Errorf("lastTrickBonus: got %d", cfg.LastTrickBonus)
		}
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.JassCpuDifficultyHard)
		target := 500
		bonus := 0
		enable := false
		p := controller.JassWebInput{
			Config: &controller.JassWebConfig{
				CpuDifficulty:  &cpu,
				TargetScore:    &target,
				LastTrickBonus: &bonus,
				EnableWeis:     &enable,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.JassCpuDifficultyHard {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 500 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
		if cfg.LastTrickBonus != 0 {
			t.Errorf("bonus: got %d", cfg.LastTrickBonus)
		}
		if cfg.EnableWeis {
			t.Errorf("enable: should be false")
		}
	})
}
