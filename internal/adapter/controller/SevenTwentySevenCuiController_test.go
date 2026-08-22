//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	ucmock "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newS27CuiController() (*controller.SevenTwentySevenCuiController, *ucmock.MockSevenTwentySevenInteractor) {
	m := new(ucmock.MockSevenTwentySevenInteractor)
	return controller.NewSevenTwentySevenCuiController(m), m
}

// **`c` は引く、`s` は止まる。** Guts の `i`/`o` を残すと、打てるのに何も
// 起きないコマンドになる。
func TestSevenTwentySevenCuiController_CardAndStand(t *testing.T) {
	for _, tt := range []struct {
		cmd  string
		draw bool
	}{{"c", true}, {"card", true}, {"s", false}, {"stand", false}} {
		t.Run(tt.cmd, func(t *testing.T) {
			c, m := newS27CuiController()
			m.On("TakeCard", tt.draw).Return("ok")
			assert.Equal(t, "ok", c.Exec(tt.cmd))
			m.AssertCalled(t, "TakeCard", tt.draw)
		})
	}
}

// **入札系のコマンドは受け付けない。** Guts から来た i/o は存在しない。
func TestSevenTwentySevenCuiController_RejectsTheGutsCommands(t *testing.T) {
	for _, cmd := range []string{"i", "in", "o", "out"} {
		t.Run(cmd, func(t *testing.T) {
			c, m := newS27CuiController()
			out := c.Exec(cmd)
			assert.NotEmpty(t, out)
			m.AssertNotCalled(t, "TakeCard", mock.Anything)
		})
	}
}

func TestSevenTwentySevenCuiController_NextRound(t *testing.T) {
	c, m := newS27CuiController()
	m.On("NextRound").Return("next")
	assert.Equal(t, "next", c.Exec("n"))
	m.AssertCalled(t, "NextRound")
}

func TestSevenTwentySevenCuiController_HintAndLog(t *testing.T) {
	c, m := newS27CuiController()
	m.On("Hint").Return("hint")
	m.On("ActionLog").Return("log")
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("l"))
}

// **設定コマンドは reset にたたむ。** config は reset でしか受け付けられないので、
// sp/sa/sc/st は現在の設定を書き換えて ResetWithConfig を呼ぶ。
func TestSevenTwentySevenCuiController_SettingsResetWithTheNewValue(t *testing.T) {
	for _, tt := range []struct {
		cmd    string
		arg    string
		expect func(domain.SevenTwentySevenConfig) bool
		what   string
	}{
		{"sp", "5", func(c domain.SevenTwentySevenConfig) bool { return c.PlayerCount == 5 }, "playerCount"},
		{"setplayers", "3", func(c domain.SevenTwentySevenConfig) bool { return c.PlayerCount == 3 }, "playerCount"},
		{"sa", "25", func(c domain.SevenTwentySevenConfig) bool { return c.Ante == 25 }, "ante"},
		{"setante", "5", func(c domain.SevenTwentySevenConfig) bool { return c.Ante == 5 }, "ante"},
		{"sc", "500", func(c domain.SevenTwentySevenConfig) bool { return c.StartingChips == 500 }, "startingChips"},
		{"setchips", "100", func(c domain.SevenTwentySevenConfig) bool { return c.StartingChips == 100 }, "startingChips"},
		{"st", "20", func(c domain.SevenTwentySevenConfig) bool { return c.TargetRounds == 20 }, "targetRounds"},
		{"setrounds", "5", func(c domain.SevenTwentySevenConfig) bool { return c.TargetRounds == 5 }, "targetRounds"},
	} {
		t.Run(tt.cmd, func(t *testing.T) {
			c, m := newS27CuiController()
			m.On("GetConfig").Return(domain.DefaultSevenTwentySevenConfig())
			m.On("ResetWithConfig", mock.MatchedBy(tt.expect)).Return("ok")
			assert.Equal(t, "ok", c.Exec(tt.cmd+" "+tt.arg))
			m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(tt.expect))
		})
	}
}

// **範囲外は弾いて reset しない。** 通すとサーバ側で黙って丸められ、
// 「設定したのに効いていない」画面になる。
func TestSevenTwentySevenCuiController_RejectsOutOfRangeSettings(t *testing.T) {
	for _, input := range []string{"sp 1", "sp 8", "sp", "sp x", "sa 0", "sc 9", "st 0"} {
		t.Run(input, func(t *testing.T) {
			c, m := newS27CuiController()
			m.On("GetConfig").Return(domain.DefaultSevenTwentySevenConfig())
			out := c.Exec(input)
			assert.NotEmpty(t, out)
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		})
	}
}

func TestSevenTwentySevenCuiController_Reset(t *testing.T) {
	c, m := newS27CuiController()
	m.On("GetConfig").Return(domain.DefaultSevenTwentySevenConfig())
	m.On("ResetWithConfig", mock.Anything).Return("reset")
	assert.Equal(t, "reset", c.Exec("r"))
}

func TestSevenTwentySevenCuiController_UnknownCommand(t *testing.T) {
	c, m := newS27CuiController()
	out := c.Exec("zzzz")
	assert.NotEmpty(t, out)
	m.AssertNotCalled(t, "TakeCard", mock.Anything)
}
