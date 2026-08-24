//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustCoincheOutputJSON(msg string) string {
	out := &controller.CoincheWebOutput{
		Players:       []*controller.CoincheWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCoincheOutputJSON: %v", err))
	}
	return string(b)
}

func TestCoincheWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockCoincheInteractor)
	biMock.On("ResetWithConfig", domain.DefaultCoincheConfig()).Return(mockOutput)
	biMock.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
	biMock.On("Coinche").Return(mockOutput)
	biMock.On("Surcoinche").Return(mockOutput)
	biMock.On("DeclineDouble").Return(mockOutput)
	biMock.On("Pass").Return(mockOutput)
	biMock.On("Play", 3).Return(mockOutput)
	biMock.On("NextTrick").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CoincheInteractorIF { return biMock }
	ctrl := controller.NewCoincheWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.CoincheWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustCoincheOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.CoincheWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("bid b", func(t *testing.T) {
		points, suit := 110, 2
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Points:       &points,
			Suit:         &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Bid", 110, 2)
	})

	// **点とスートは 2 つで 1 つの宣言。** 片方だけ来た要求を通すと、残りに
	// 既定値が入って別の契約になる。
	t.Run("bid needs both halves", func(t *testing.T) {
		points, suit := 110, 2
		for _, in := range []controller.CoincheWebInput{
			{BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "s1"}},
			{BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "s1"}, Points: &points},
			{BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "s1"}, Suit: &suit},
		} {
			input := in
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("doubling commands", func(t *testing.T) {
		for _, cmd := range []string{"co", "coinche", "su", "surcoinche", "ok", "decline"} {
			input := controller.CoincheWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "s1"},
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("pass pa", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pa", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play p with index", func(t *testing.T) {
		idx := 3
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing index", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("next n", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround nr", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		input := controller.CoincheWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestCoincheWebConfig_ToConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		p := controller.CoincheWebInput{}
		cfg := p.ToConfig()
		assert := func(want, got int, name string) {
			if want != got {
				t.Errorf("%s: want %d, got %d", name, want, got)
			}
		}
		assert(int(domain.CoincheCpuDifficultyNormal), int(cfg.CpuDifficulty), "difficulty")
		assert(1000, cfg.TargetScore, "target")
		assert(10, cfg.DixDeDer, "dix")
	})

	t.Run("explicit config", func(t *testing.T) {
		cpu := int(domain.CoincheCpuDifficultyHard)
		target := 500
		dix := 0
		enable := false
		p := controller.CoincheWebInput{
			Config: &controller.CoincheWebConfig{
				CpuDifficulty:        &cpu,
				TargetScore:          &target,
				DixDeDer:             &dix,
				EnableBeloteRebelote: &enable,
			},
		}
		cfg := p.ToConfig()
		if cfg.CpuDifficulty != domain.CoincheCpuDifficultyHard {
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
