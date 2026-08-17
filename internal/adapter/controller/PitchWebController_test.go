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

func mustPitchOutputJSON(msg string) string {
	out := &controller.PitchWebOutput{
		Players:         []*controller.PitchWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrick:       []*controller.WebOutputTrickCard{},
		LastTrickWinner: -1,
		WinnerIdx:       -1,
		BidWinnerIdx:    -1,
		// まだ何も争われていないので、どのカテゴリも「なし」(#5584)。
		RoundBreakdown: &controller.PitchWebOutputBreakdown{
			High: domain.PitchNoScorer,
			Low:  domain.PitchNoScorer,
			Jack: domain.PitchNoScorer,
			Game: domain.PitchNoScorer,
		},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPitchOutputJSON: %v", err))
	}
	return string(b)
}

func TestPitchWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	piMock := new(usecase.MockPitchInteractor)
	piMock.On("ResetWithConfig", domain.DefaultPitchConfig()).Return(mockOutput)
	piMock.On("Bid", 3).Return(mockOutput)
	piMock.On("Bid", 0).Return(mockOutput)
	piMock.On("Play", 2).Return(mockOutput)
	piMock.On("NextTrick").Return(mockOutput)
	piMock.On("NextRound").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PitchInteractorIF { return piMock }
	ctrl := controller.NewPitchWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q returns bye", func(t *testing.T) {
		var input controller.PitchWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustPitchOutputJSON("bye."))
	})
	t.Run("reset r", func(t *testing.T) {
		var input controller.PitchWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid 3", func(t *testing.T) {
		bid := 3
		input := controller.PitchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          &bid,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid pass", func(t *testing.T) {
		bid := 0
		input := controller.PitchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "s1"},
			Bid:          &bid,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing param 400", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("play 2", func(t *testing.T) {
		idx := 2
		input := controller.PitchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex 400", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("next n", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("nextround nr", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("hint h", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("log l", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "l", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("unknown command", func(t *testing.T) {
		input := controller.PitchWebInput{BaseWebInput: controller.BaseWebInput{Command: "xyz", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestPitchWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		var p controller.PitchWebInput
		cfg := p.ToConfig()
		if cfg != domain.DefaultPitchConfig() {
			t.Fatalf("expected default config, got %+v", cfg)
		}
	})
	t.Run("provided config bound-checks", func(t *testing.T) {
		diff := 999
		limit := -5
		p := controller.PitchWebInput{Config: &controller.PitchWebConfig{CpuDifficulty: &diff, PointLimit: &limit}}
		cfg := p.ToConfig()
		if int(cfg.CpuDifficulty) > int(domain.PitchCpuDifficultyHard) {
			t.Fatalf("CpuDifficulty out of range: %v", cfg.CpuDifficulty)
		}
		if cfg.PointLimit < 1 {
			t.Fatalf("PointLimit out of range: %v", cfg.PointLimit)
		}
	})
}
