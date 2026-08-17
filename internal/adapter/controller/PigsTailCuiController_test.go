package controller_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPigsTailCuiController_Method(t *testing.T) {
	mockOutput := "==========\nPig's Tail (ぶたのしっぽ)\n=========="
	ptiMock := new(usecase.MockPigsTailInteractor)
	ptiMock.On("GetConfig").Return(domain.DefaultPigsTailConfig())
	ptiMock.On("Reset", mock.Anything).Return(mockOutput)
	ptiMock.On("Action", 0).Return(mockOutput)
	ptiMock.On("ActionLog").Return(`[]`)

	tptc := controller.NewPigsTailCuiController(ptiMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tptc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tptc.Exec("quit"))
	})
	t.Run("success Exec r preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("r"))
		ptiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec reset preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("reset"))
		ptiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("d"))
		ptiMock.AssertCalled(t, "Action", 0)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("draw"))
		ptiMock.AssertCalled(t, "Action", 0)
	})
	t.Run("success Exec log", func(t *testing.T) {
		assert.Equal(t, `[]`, tptc.Exec("log"))
		ptiMock.AssertCalled(t, "ActionLog")
	})
	t.Run("success Exec other returns unknown", func(t *testing.T) {
		result := tptc.Exec("xyz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}

// #5521: Web は 2〜6 人を選べるのに、CUI には人数を変える手段が一つも無く、
// 常に既定の 4 人で始まっていた。
func TestPigsTailCuiController_SetPlayers(t *testing.T) {
	newCtrl := func() (*controller.PigsTailCuiController, *usecase.MockPigsTailInteractor) {
		m := new(usecase.MockPigsTailInteractor)
		// **他の設定は残ること。**新しい既定値から作り直すと、CPU の
		// 迷い時間の設定が人数変更のたびに消える。
		m.On("GetConfig").Return(domain.PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: 4})
		m.On("Reset", mock.Anything).Return("reset output")
		return controller.NewPigsTailCuiController(m), m
	}

	for _, cmd := range []string{"sp", "setplayers"} {
		t.Run("sets the count via "+cmd, func(t *testing.T) {
			c, m := newCtrl()
			assert.Equal(t, "reset output", c.Exec(cmd+" 6"))
			m.AssertCalled(t, "Reset", domain.PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: 6})
		})
	}

	// **範囲外は設定を変えずに断る。**通してしまうと、次の reset で
	// 人数不正のまま配ることになる。
	for _, arg := range []string{"1", "7", "abc", ""} {
		t.Run("rejects "+arg, func(t *testing.T) {
			c, m := newCtrl()
			out := c.Exec(strings.TrimSpace("sp " + arg))
			assert.NotEqual(t, "reset output", out)
			assert.NotEmpty(t, out)
			m.AssertNotCalled(t, "Reset", mock.Anything)
		})
	}

	// 打ち間違えたときに候補として出ること。既知コマンド一覧に載せ忘れても
	// sp 自体は動いてしまうので、ここを見ないと抜けに気付けない。
	t.Run("is offered as a suggestion for a near miss", func(t *testing.T) {
		c, _ := newCtrl()
		assert.Contains(t, c.Exec("setplayer"), "setplayers")
	})

	t.Run("the boundaries are inside the range", func(t *testing.T) {
		for _, n := range []int{domain.PigsTailMinPlayers, domain.PigsTailMaxPlayers} {
			c, m := newCtrl()
			assert.Equal(t, "reset output", c.Exec("sp "+strconv.Itoa(n)))
			m.AssertCalled(t, "Reset", domain.PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: n})
		}
	})
}
