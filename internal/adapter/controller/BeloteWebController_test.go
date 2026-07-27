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

func mustBeloteOutputJSON(msg string) string {
	out := &controller.BeloteWebOutput{
		Players:       []*controller.BeloteWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBeloteOutputJSON: %v", err))
	}
	return string(b)
}

func TestBeloteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockBeloteInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBeloteConfig()).Return(mockOutput)
	biMock.On("PickUp", true).Return(mockOutput)
	biMock.On("CallTrump", 2).Return(mockOutput)
	biMock.On("Pass").Return(mockOutput)
	biMock.On("Play", 3).Return(mockOutput)
	biMock.On("NextTrick").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BeloteInteractorIF { return biMock }
	ctrl := controller.NewBeloteWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.BeloteWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustBeloteOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BeloteWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("orderup o", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "o", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("calltrump c", func(t *testing.T) {
		suit := 2
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "s1"},
			Suit:         &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("calltrump missing suit", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "calltrump", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("pass pa", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pa", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play p with index", func(t *testing.T) {
		idx := 3
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("next n", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround nr", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.BeloteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestBeloteWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.BeloteWebInput{}
		cfg := p.ToConfig()
		assert := func(want, got int, name string) {
			if want != got {
				t.Errorf("%s: want %d, got %d", name, want, got)
			}
		}
		assert(int(domain.BeloteCpuDifficultyNormal), int(cfg.CpuDifficulty), "difficulty")
		assert(1000, cfg.TargetScore, "target")
		assert(10, cfg.DixDeDer, "dix")
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.BeloteCpuDifficultyHard)
		target := 500
		dix := 0
		enable := false
		p := controller.BeloteWebInput{
			Config: &controller.BeloteWebConfig{
				CpuDifficulty:        &cpu,
				TargetScore:          &target,
				DixDeDer:             &dix,
				EnableBeloteRebelote: &enable,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.BeloteCpuDifficultyHard {
			t.Errorf("difficulty: got %d", cfg.CpuDifficulty)
		}
		if cfg.TargetScore != 500 {
			t.Errorf("target: got %d", cfg.TargetScore)
		}
		if cfg.DixDeDer != 0 {
			t.Errorf("dix: got %d", cfg.DixDeDer)
		}
		if cfg.EnableBeloteRebelote {
			t.Errorf("enable: should be false")
		}
	})
}
