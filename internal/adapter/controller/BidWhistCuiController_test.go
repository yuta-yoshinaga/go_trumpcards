//go:build test

package controller_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBidWhistCui() (*controller.BidWhistCuiController, *usecase.MockBidWhistInteractor) {
	biMock := new(usecase.MockBidWhistInteractor)
	biMock.On("GetConfig").Return(domain.DefaultBidWhistConfig())
	biMock.On("ResetWithConfig", domain.DefaultBidWhistConfig()).Return("reset")
	biMock.On("Bid", 4, domain.BidWhistDirectionUptown).Return("bid")
	biMock.On("Bid", 5, domain.BidWhistDirectionNoTrump).Return("bid-nt")
	biMock.On("Pass").Return("pass")
	biMock.On("DeclareTrump", 1).Return("trump")
	biMock.On("ExchangeKitty", []int{0, 1, 2, 3, 4, 5}).Return("exchange")
	biMock.On("Play", 3).Return("play")
	biMock.On("NextTrick").Return("next")
	biMock.On("NextRound").Return("nextround")
	biMock.On("Hint").Return("hint")
	biMock.On("ActionLog").Return("log")
	return controller.NewBidWhistCuiController(biMock), biMock
}

func TestBidWhistCuiController_Exec(t *testing.T) {
	c, _ := newBidWhistCui()
	cases := []struct {
		cmd  string
		want string
	}{
		{"r", "reset"},
		{"b 4 0", "bid"},
		{"bid 4 0", "bid"},
		{"b 5 2", "bid-nt"},
		{"pa", "pass"},
		{"t 1", "trump"},
		{"e 0 1 2 3 4 5", "exchange"},
		{"p 3", "play"},
		{"n", "next"},
		{"nr", "nextround"},
		{"h", "hint"},
		{"log", "log"},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			got := c.Exec(tc.cmd)
			if !strings.Contains(got, tc.want) {
				t.Errorf("Exec(%q) = %q, want contains %q", tc.cmd, got, tc.want)
			}
		})
	}
}

func TestBidWhistCuiController_Quit(t *testing.T) {
	c, _ := newBidWhistCui()
	if got := c.Exec("q"); !strings.Contains(got, "bye") {
		t.Errorf("quit = %q, want bye", got)
	}
}

func TestBidWhistCuiController_Usage(t *testing.T) {
	c, _ := newBidWhistCui()
	if got := c.Exec("b"); !msgRejected(got) {
		t.Errorf("bid without args should show usage, got %q", got)
	}
	if got := c.Exec("e 0 1 2"); !msgRejected(got) {
		t.Errorf("exchange with too few args should show usage, got %q", got)
	}
	if got := c.Exec("p"); !strings.Contains(got, msgCardIndexRequired()) {
		t.Errorf("play without args should require index, got %q", got)
	}
	if got := c.Exec("t"); !strings.Contains(got, msgStem("trumpSuitRequiredLettersPlain")) {
		t.Errorf("trump without args should require suit, got %q", got)
	}
}

func TestBidWhistCuiController_Settings(t *testing.T) {
	biMock := new(usecase.MockBidWhistInteractor)
	cfg := domain.DefaultBidWhistConfig()
	biMock.On("GetConfig").Return(cfg)
	biMock.On("ResetWithConfig", domain.BidWhistConfig{CpuDifficulty: domain.BidWhistCpuDifficultyHard, TargetScore: 7}).Return("sd")
	biMock.On("ResetWithConfig", domain.BidWhistConfig{CpuDifficulty: domain.BidWhistCpuDifficultyNormal, TargetScore: 9}).Return("st")
	c := controller.NewBidWhistCuiController(biMock)
	if got := c.Exec("sd 2"); !strings.Contains(got, "sd") {
		t.Errorf("setdifficulty = %q", got)
	}
	if got := c.Exec("st 9"); !strings.Contains(got, "st") {
		t.Errorf("settarget = %q", got)
	}
}
