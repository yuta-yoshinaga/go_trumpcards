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

func newFiveHundredMockController(t *testing.T) (*controller.FiveHundredWebController, *usecase.MockFiveHundredInteractor) {
	t.Helper()
	mockOutput := `{"ok":true}`
	fiMock := new(usecase.MockFiveHundredInteractor)
	fiMock.On("ResetWithConfig", domain.DefaultFiveHundredConfig()).Return(mockOutput)
	fiMock.On("Bid", domain.FiveHundredContractSuit, 7, 1).Return(mockOutput)
	fiMock.On("Bid", domain.FiveHundredContractMisere, 0, -1).Return(mockOutput)
	fiMock.On("Pass").Return(mockOutput)
	fiMock.On("ExchangeKitty", []int{0, 1, 2}).Return(mockOutput)
	fiMock.On("Play", 3, -1).Return(mockOutput)
	fiMock.On("Play", 3, 2).Return(mockOutput)
	fiMock.On("NextTrick").Return(mockOutput)
	fiMock.On("NextRound").Return(mockOutput)
	fiMock.On("Hint").Return(mockOutput)
	fiMock.On("ActionLog").Return(mockOutput)
	ctrl := controller.NewFiveHundredWebController(func() uc.FiveHundredInteractorIF { return fiMock })
	t.Cleanup(ctrl.Stop)
	return ctrl, fiMock
}

func intPtr500(v int) *int { return &v }

func TestFiveHundredWebController_Commands(t *testing.T) {
	ctrl, fiMock := newFiveHundredMockController(t)

	cases := []struct {
		name  string
		input controller.FiveHundredWebInput
		want  string
	}{
		{"reset", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "r"}}, `{"ok":true}`},
		{"bid suit", controller.FiveHundredWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b"},
			BidKind:      intPtr500(int(domain.FiveHundredContractSuit)), BidTricks: intPtr500(7), BidSuit: intPtr500(1),
		}, `{"ok":true}`},
		{"bid misere", controller.FiveHundredWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid"},
			BidKind:      intPtr500(int(domain.FiveHundredContractMisere)),
		}, `{"ok":true}`},
		{"pass", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "pa"}}, `{"ok":true}`},
		{"exchange", controller.FiveHundredWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "e"}, DiscardIndices: []int{0, 1, 2},
		}, `{"ok":true}`},
		{"play", controller.FiveHundredWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p"}, CardIndex: intPtr500(3),
		}, `{"ok":true}`},
		{"play joker suit", controller.FiveHundredWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play"}, CardIndex: intPtr500(3), JokerSuit: intPtr500(2),
		}, `{"ok":true}`},
		{"next", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "n"}}, `{"ok":true}`},
		{"nextround", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "nr"}}, `{"ok":true}`},
		{"hint", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "hint"}}, `{"ok":true}`},
		{"log", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "log"}}, `{"ok":true}`},
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
	fiMock.AssertCalled(t, "Bid", domain.FiveHundredContractSuit, 7, 1)
	fiMock.AssertCalled(t, "ExchangeKitty", []int{0, 1, 2})
	fiMock.AssertCalled(t, "Play", 3, 2)
}

func TestFiveHundredWebController_MissingParams(t *testing.T) {
	ctrl, _ := newFiveHundredMockController(t)
	cases := []struct {
		name  string
		input controller.FiveHundredWebInput
	}{
		{"bid missing kind", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "b"}}},
		{"exchange missing", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "e"}}},
		{"play missing", controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "p"}}},
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

func TestFiveHundredWebController_Quit(t *testing.T) {
	ctrl, _ := newFiveHundredMockController(t)
	input := controller.FiveHundredWebInput{BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "s1"}}
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusOK)
	if !strings.Contains(rec.Body.String(), "bye") {
		t.Errorf("expected bye message, got %s", rec.Body.String())
	}
}

func TestFiveHundredWebConfig_ToConfig(t *testing.T) {
	cd := 2
	ts := 300
	in := controller.FiveHundredWebInput{Config: &controller.FiveHundredWebConfig{CpuDifficulty: &cd, TargetScore: &ts}}
	cfg := in.ToConfig()
	if cfg.CpuDifficulty != domain.FiveHundredCpuDifficultyHard {
		t.Errorf("cpuDifficulty = %d, want Hard", cfg.CpuDifficulty)
	}
	if cfg.TargetScore != 300 {
		t.Errorf("targetScore = %d, want 300", cfg.TargetScore)
	}
	// nil config falls back to default.
	def := controller.FiveHundredWebInput{}.ToConfig()
	if def.TargetScore != domain.FiveHundredDefaultTargetScore {
		t.Errorf("default targetScore = %d, want %d", def.TargetScore, domain.FiveHundredDefaultTargetScore)
	}
}
