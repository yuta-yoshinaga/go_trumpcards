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

func newBidWhistMockController(t *testing.T) (*controller.BidWhistWebController, *usecase.MockBidWhistInteractor) {
	t.Helper()
	mockOutput := `{"ok":true}`
	biMock := new(usecase.MockBidWhistInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBidWhistConfig()).Return(mockOutput)
	biMock.On("Bid", 4, domain.BidWhistDirectionUptown).Return(mockOutput)
	biMock.On("Pass").Return(mockOutput)
	biMock.On("DeclareTrump", 1).Return(mockOutput)
	biMock.On("ExchangeKitty", []int{0, 1, 2, 3, 4, 5}).Return(mockOutput)
	biMock.On("Play", 3).Return(mockOutput)
	biMock.On("NextTrick").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)
	ctrl := controller.NewBidWhistWebController(func() uc.BidWhistInteractorIF { return biMock })
	t.Cleanup(ctrl.Stop)
	return ctrl, biMock
}

func intPtrBW(v int) *int { return &v }

func TestBidWhistWebController_Commands(t *testing.T) {
	ctrl, biMock := newBidWhistMockController(t)

	cases := []struct {
		name  string
		input controller.BidWhistWebInput
		want  string
	}{
		{"reset", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "r"}}, `{"ok":true}`},
		{"bid", controller.BidWhistWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b"},
			BidTricks:    intPtrBW(4), BidDirection: intPtrBW(domain.BidWhistDirectionUptown),
		}, `{"ok":true}`},
		{"pass", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "pa"}}, `{"ok":true}`},
		{"trump", controller.BidWhistWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "t"}, TrumpSuit: intPtrBW(1),
		}, `{"ok":true}`},
		{"exchange", controller.BidWhistWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "e"}, DiscardIndices: []int{0, 1, 2, 3, 4, 5},
		}, `{"ok":true}`},
		{"play", controller.BidWhistWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p"}, CardIndex: intPtrBW(3),
		}, `{"ok":true}`},
		{"next", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "n"}}, `{"ok":true}`},
		{"nextround", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "nr"}}, `{"ok":true}`},
		{"hint", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "hint"}}, `{"ok":true}`},
		{"log", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "log"}}, `{"ok":true}`},
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
	biMock.AssertCalled(t, "Bid", 4, domain.BidWhistDirectionUptown)
	biMock.AssertCalled(t, "DeclareTrump", 1)
	biMock.AssertCalled(t, "ExchangeKitty", []int{0, 1, 2, 3, 4, 5})
	biMock.AssertCalled(t, "Play", 3)
}

func TestBidWhistWebController_MissingParams(t *testing.T) {
	ctrl, _ := newBidWhistMockController(t)
	cases := []struct {
		name  string
		input controller.BidWhistWebInput
	}{
		{"bid missing", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "b"}}},
		{"trump missing", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "t"}}},
		{"exchange missing", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "e"}}},
		{"play missing", controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "p"}}},
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

func TestBidWhistWebController_Quit(t *testing.T) {
	ctrl, _ := newBidWhistMockController(t)
	input := controller.BidWhistWebInput{BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "s1"}}
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusOK)
	if !strings.Contains(rec.Body.String(), "bye") {
		t.Errorf("expected bye message, got %s", rec.Body.String())
	}
}

func TestBidWhistWebConfig_ToConfig(t *testing.T) {
	cd := 2
	ts := 9
	in := controller.BidWhistWebInput{Config: &controller.BidWhistWebConfig{CpuDifficulty: &cd, TargetScore: &ts}}
	cfg := in.ToConfig()
	if cfg.CpuDifficulty != domain.BidWhistCpuDifficultyHard {
		t.Errorf("cpuDifficulty = %d, want Hard", cfg.CpuDifficulty)
	}
	if cfg.TargetScore != 9 {
		t.Errorf("targetScore = %d, want 9", cfg.TargetScore)
	}
	def := controller.BidWhistWebInput{}.ToConfig()
	if def.TargetScore != domain.BidWhistDefaultTargetScore {
		t.Errorf("default targetScore = %d, want %d", def.TargetScore, domain.BidWhistDefaultTargetScore)
	}
}
