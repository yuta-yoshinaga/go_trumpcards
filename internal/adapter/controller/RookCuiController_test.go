//go:build test

package controller_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newRookCui() (*controller.RookCuiController, *usecase.MockRookInteractor) {
	fiMock := new(usecase.MockRookInteractor)
	fiMock.On("GetConfig").Return(domain.DefaultRookConfig())
	fiMock.On("ResetWithConfig", domain.DefaultRookConfig()).Return("reset")
	fiMock.On("Bid", 75).Return("bid")
	fiMock.On("Pass").Return("pass")
	fiMock.On("ExchangeNest", []int{0, 1, 2, 3, 4}, 3).Return("exchange")
	fiMock.On("Play", 3).Return("play")
	fiMock.On("NextTrick").Return("next")
	fiMock.On("NextRound").Return("nextround")
	fiMock.On("Hint").Return("hint")
	fiMock.On("ActionLog").Return("log")
	return controller.NewRookCuiController(fiMock), fiMock
}

func TestRookCuiController_Exec(t *testing.T) {
	c, _ := newRookCui()
	cases := []struct {
		cmd  string
		want string
	}{
		{"r", "reset"},
		{"b 75", "bid"},
		{"bid 75", "bid"},
		{"pa", "pass"},
		{"e 0 1 2 3 4 3", "exchange"},
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

func TestRookCuiController_Quit(t *testing.T) {
	c, _ := newRookCui()
	if got := c.Exec("q"); !strings.Contains(got, "bye") {
		t.Errorf("quit = %q, want bye", got)
	}
}

func TestRookCuiController_Usage(t *testing.T) {
	c, _ := newRookCui()
	if got := c.Exec("b"); !strings.Contains(got, msgStem("bidRequired70120")) {
		t.Errorf("bid without args should require, got %q", got)
	}
	if got := c.Exec("e 0 1 2"); !msgRejected(got) {
		t.Errorf("exchange with few args should show usage, got %q", got)
	}
	if got := c.Exec("p"); !strings.Contains(got, msgCardIndexRequired()) {
		t.Errorf("play without args should require index, got %q", got)
	}
}

func TestRookCuiController_Settings(t *testing.T) {
	fiMock := new(usecase.MockRookInteractor)
	cfg := domain.DefaultRookConfig()
	fiMock.On("GetConfig").Return(cfg)
	fiMock.On("ResetWithConfig", domain.RookConfig{CpuDifficulty: domain.RookCpuDifficultyHard, TargetScore: 500}).Return("sd")
	fiMock.On("ResetWithConfig", domain.RookConfig{CpuDifficulty: domain.RookCpuDifficultyNormal, TargetScore: 300}).Return("st")
	c := controller.NewRookCuiController(fiMock)
	if got := c.Exec("sd 2"); !strings.Contains(got, "sd") {
		t.Errorf("setdifficulty = %q", got)
	}
	if got := c.Exec("st 300"); !strings.Contains(got, "st") {
		t.Errorf("settarget = %q", got)
	}
}
