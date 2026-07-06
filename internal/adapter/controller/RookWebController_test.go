//go:build test

package controller_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newRookMockController(t *testing.T) (*controller.RookWebController, *usecase.MockRookInteractor) {
	t.Helper()
	mockOutput := `{"ok":true}`
	fiMock := new(usecase.MockRookInteractor)
	fiMock.On("ResetWithConfig", domain.DefaultRookConfig()).Return(mockOutput)
	fiMock.On("Bid", 75).Return(mockOutput)
	fiMock.On("Pass").Return(mockOutput)
	fiMock.On("ExchangeNest", []int{0, 1, 2, 3, 4}, 3).Return(mockOutput)
	fiMock.On("Play", 3).Return(mockOutput)
	fiMock.On("NextTrick").Return(mockOutput)
	fiMock.On("NextRound").Return(mockOutput)
	fiMock.On("Hint").Return(mockOutput)
	fiMock.On("ActionLog").Return(mockOutput)
	ctrl := controller.NewRookWebController(func() uc.RookInteractorIF { return fiMock })
	t.Cleanup(ctrl.Stop)
	return ctrl, fiMock
}

func intPtrRook(v int) *int { return &v }

func TestRookWebController_Commands(t *testing.T) {
	ctrl, fiMock := newRookMockController(t)

	cases := []struct {
		name  string
		input controller.RookWebInput
		want  string
	}{
		{"reset", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "r"}}, `{"ok":true}`},
		{"bid", controller.RookWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b"}, Bid: intPtrRook(75),
		}, `{"ok":true}`},
		{"pass", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "pa"}}, `{"ok":true}`},
		{"exchange", controller.RookWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "e"}, DiscardIndices: []int{0, 1, 2, 3, 4}, TrumpColor: intPtrRook(3),
		}, `{"ok":true}`},
		{"play", controller.RookWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p"}, CardIndex: intPtrRook(3),
		}, `{"ok":true}`},
		{"next", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "n"}}, `{"ok":true}`},
		{"nextround", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "nr"}}, `{"ok":true}`},
		{"hint", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "hint"}}, `{"ok":true}`},
		{"log", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "log"}}, `{"ok":true}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.input.SessionID = "s1"
			rec := execRequest(t, ctrl.Exec, &c.input)
			rec.CodeIs(http.StatusOK)
			rec.ContentTypeIsJson()
			rec.BodyIs(c.want)
		})
	}
	fiMock.AssertCalled(t, "Bid", 75)
	fiMock.AssertCalled(t, "ExchangeNest", []int{0, 1, 2, 3, 4}, 3)
	fiMock.AssertCalled(t, "Play", 3)
}

func TestRookWebController_MissingParams(t *testing.T) {
	ctrl, _ := newRookMockController(t)
	cases := []struct {
		name  string
		input controller.RookWebInput
	}{
		{"bid missing", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "b"}}},
		{"exchange missing discards", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "e"}, TrumpColor: intPtrRook(1)}},
		{"exchange missing trump", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "e"}, DiscardIndices: []int{0, 1, 2, 3, 4}}},
		{"play missing", controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "p"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.input.SessionID = "s1"
			rec := execRequest(t, ctrl.Exec, &c.input)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), "is required") {
				t.Errorf("expected 'is required' message, got %s", rec.Body.String())
			}
		})
	}
}

func TestRookWebController_Quit(t *testing.T) {
	ctrl, _ := newRookMockController(t)
	input := controller.RookWebInput{BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "s1"}}
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusOK)
	if !strings.Contains(rec.Body.String(), "bye") {
		t.Errorf("expected bye message, got %s", rec.Body.String())
	}
}

func TestRookWebConfig_ToConfig(t *testing.T) {
	cd := 2
	ts := 300
	in := controller.RookWebInput{Config: &controller.RookWebConfig{CpuDifficulty: &cd, TargetScore: &ts}}
	cfg := in.ToConfig()
	if cfg.CpuDifficulty != domain.RookCpuDifficultyHard {
		t.Errorf("cpuDifficulty = %d, want Hard", cfg.CpuDifficulty)
	}
	if cfg.TargetScore != 300 {
		t.Errorf("targetScore = %d, want 300", cfg.TargetScore)
	}
	def := controller.RookWebInput{}.ToConfig()
	if def.TargetScore != domain.RookDefaultTargetScore {
		t.Errorf("default targetScore = %d, want %d", def.TargetScore, domain.RookDefaultTargetScore)
	}
}
