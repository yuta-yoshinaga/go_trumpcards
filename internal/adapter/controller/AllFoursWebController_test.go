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

func mustAllFoursOutputJSON(msg string) string {
	out := &controller.AllFoursWebOutput{
		Players:       []*controller.AllFoursWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAllFoursOutputJSON: %v", err))
	}
	return string(b)
}

func TestAllFoursWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	aiMock := new(usecase.MockAllFoursInteractor)
	aiMock.On("ResetWithConfig", domain.DefaultAllFoursConfig()).Return(mockOutput)
	aiMock.On("Beg", true).Return(mockOutput)
	aiMock.On("Beg", false).Return(mockOutput)
	aiMock.On("RespondBeg", true).Return(mockOutput)
	aiMock.On("RespondBeg", false).Return(mockOutput)
	aiMock.On("Play", 2).Return(mockOutput)
	aiMock.On("NextTrick").Return(mockOutput)
	aiMock.On("NextRound").Return(mockOutput)
	aiMock.On("Hint").Return(mockOutput)
	aiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.AllFoursInteractorIF { return aiMock }
	ctrl := controller.NewAllFoursWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q returns bye", func(t *testing.T) {
		var input controller.AllFoursWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustAllFoursOutputJSON("bye."))
	})
	t.Run("reset r", func(t *testing.T) {
		var input controller.AllFoursWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("beg true", func(t *testing.T) {
		beg := true
		input := controller.AllFoursWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "beg", SessionID: "s1"},
			Beg:          &beg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("beg missing param 400", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "beg", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("respond run", func(t *testing.T) {
		run := true
		input := controller.AllFoursWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "respond", SessionID: "s1"},
			Run:          &run,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("respond missing param 400", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "respond", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("play 2", func(t *testing.T) {
		idx := 2
		input := controller.AllFoursWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex 400", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("next n", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("nextround nr", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "nr", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("hint h", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("log l", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "l", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("unknown command", func(t *testing.T) {
		input := controller.AllFoursWebInput{BaseWebInput: controller.BaseWebInput{Command: "xyz", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestAllFoursWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		var in controller.AllFoursWebInput
		cfg := in.ToConfig()
		if cfg != domain.DefaultAllFoursConfig() {
			t.Fatalf("expected default config, got %+v", cfg)
		}
	})
	t.Run("provided config bound-checks", func(t *testing.T) {
		diff := 999
		limit := -5
		in := controller.AllFoursWebInput{Config: &controller.AllFoursWebConfig{CpuDifficulty: &diff, PointLimit: &limit}}
		cfg := in.ToConfig()
		if int(cfg.CpuDifficulty) > int(domain.AllFoursCpuDifficultyHard) {
			t.Fatalf("CpuDifficulty out of range: %v", cfg.CpuDifficulty)
		}
		if cfg.PointLimit < 1 {
			t.Fatalf("PointLimit out of range: %v", cfg.PointLimit)
		}
	})
}
