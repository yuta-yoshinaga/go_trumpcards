//go:build test

package controller_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newFiveHundredCui() (*controller.FiveHundredCuiController, *usecase.MockFiveHundredInteractor) {
	fiMock := new(usecase.MockFiveHundredInteractor)
	fiMock.On("GetConfig").Return(domain.DefaultFiveHundredConfig())
	fiMock.On("ResetWithConfig", domain.DefaultFiveHundredConfig()).Return("reset")
	fiMock.On("Bid", domain.FiveHundredContractSuit, 7, 1).Return("bid-suit")
	fiMock.On("Bid", domain.FiveHundredContractNoTrump, 8, -1).Return("bid-nt")
	fiMock.On("Bid", domain.FiveHundredContractMisere, 0, -1).Return("bid-misere")
	fiMock.On("Bid", domain.FiveHundredContractOpenMisere, 0, -1).Return("bid-openmisere")
	fiMock.On("Pass").Return("pass")
	fiMock.On("ExchangeKitty", []int{0, 1, 2}).Return("exchange")
	fiMock.On("Play", 3, -1).Return("play")
	fiMock.On("Play", 3, 2).Return("play-joker")
	fiMock.On("NextTrick").Return("next")
	fiMock.On("NextRound").Return("nextround")
	fiMock.On("Hint").Return("hint")
	fiMock.On("ActionLog").Return("log")
	return controller.NewFiveHundredCuiController(fiMock), fiMock
}

func TestFiveHundredCuiController_Exec(t *testing.T) {
	c, _ := newFiveHundredCui()
	cases := []struct {
		cmd  string
		want string
	}{
		{"r", "reset"},
		{"b 7 1", "bid-suit"},
		{"bid 7 1", "bid-suit"},
		{"bnt 8", "bid-nt"},
		{"m", "bid-misere"},
		{"om", "bid-openmisere"},
		{"pa", "pass"},
		{"e 0 1 2", "exchange"},
		{"p 3", "play"},
		{"p 3 2", "play-joker"},
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

func TestFiveHundredCuiController_Quit(t *testing.T) {
	c, _ := newFiveHundredCui()
	if got := c.Exec("q"); !strings.Contains(got, "bye") {
		t.Errorf("quit = %q, want bye", got)
	}
}

func TestFiveHundredCuiController_Usage(t *testing.T) {
	c, _ := newFiveHundredCui()
	if got := c.Exec("b"); !strings.Contains(got, "Usage") {
		t.Errorf("bid without args should show usage, got %q", got)
	}
	if got := c.Exec("e 0"); !strings.Contains(got, "Usage") {
		t.Errorf("exchange with one arg should show usage, got %q", got)
	}
	if got := c.Exec("p"); !strings.Contains(got, msgCardIndexRequired()) {
		t.Errorf("play without args should require index, got %q", got)
	}
}

func TestFiveHundredCuiController_Settings(t *testing.T) {
	fiMock := new(usecase.MockFiveHundredInteractor)
	cfg := domain.DefaultFiveHundredConfig()
	fiMock.On("GetConfig").Return(cfg)
	fiMock.On("ResetWithConfig", domain.FiveHundredConfig{CpuDifficulty: domain.FiveHundredCpuDifficultyHard, TargetScore: 500}).Return("sd")
	fiMock.On("ResetWithConfig", domain.FiveHundredConfig{CpuDifficulty: domain.FiveHundredCpuDifficultyNormal, TargetScore: 300}).Return("st")
	c := controller.NewFiveHundredCuiController(fiMock)
	if got := c.Exec("sd 2"); !strings.Contains(got, "sd") {
		t.Errorf("setdifficulty = %q", got)
	}
	if got := c.Exec("st 300"); !strings.Contains(got, "st") {
		t.Errorf("settarget = %q", got)
	}
}
