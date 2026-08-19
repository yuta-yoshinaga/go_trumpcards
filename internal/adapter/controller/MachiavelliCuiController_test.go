//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestMachiavelliCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockMachiavelliInteractor {
		m := new(mockUsecases.MockMachiavelliInteractor)
		m.On("GetConfig").Return(domain.DefaultMachiavelliConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Draw").Return(mockOutput)
		m.On("NewMeld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewMachiavelliCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMachiavelliConfig())
	})

	t.Run("draw aliases", func(t *testing.T) {
		for _, cmd := range []string{"dr", "draw"} {
			m := newMock()
			c := controller.NewMachiavelliCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "Draw")
		}
	})

	t.Run("newmeld", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nm 0 1 2"))
		m.AssertCalled(t, "NewMeld", []int{0, 1, 2})
	})

	t.Run("newmeld usage when too few", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("nm 0 1")
		assert.Contains(t, out, msgUsage("usageNmIJKAtLeast3HandIndices"))
		m.AssertNotCalled(t, "NewMeld", mock.Anything)
	})

	t.Run("layoff", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 0 2"))
		m.AssertCalled(t, "Layoff", 0, 2)
	})

	t.Run("layoff usage when missing args", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("lo 0")
		assert.Contains(t, out, msgUsage("usageLoMeldidxHandindex"))
		m.AssertNotCalled(t, "Layoff", mock.Anything, mock.Anything)
	})

	t.Run("nextround aliases", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			m := newMock()
			c := controller.NewMachiavelliCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "NextRound")
		}
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pc 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.PlayerCount == 5
		}))
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.CpuDifficulty == domain.MachiavelliCpuDifficultyHard
		}))
	})

	t.Run("setrounds", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sr 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.TargetRounds == 5
		}))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("blarghhh")
		assert.NotEmpty(t, out)
	})

	// #5704: 「手札を出せるなら場のメルドを自由に組み替えてよい」がこのゲームの
	// 核心ルールで、Web は丸ごと 1 機能として実装しているのに、CUI にはコマンドが
	// 無く、代表的な戦略行為を実行する手段が無かった。
	t.Run("rearrange rebuilds the table and plays from hand", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)

		assert.Equal(t, mockOutput, c.Exec("ra s5,h5,d5;c7,c8,c9 / 2,4"))

		m.AssertCalled(t, "Play", [][]domain.MachiavelliCardRef{
			{
				{Design: domain.CardDesignSpade, Value: 5},
				{Design: domain.CardDesignHeart, Value: 5},
				{Design: domain.CardDesignDiamond, Value: 5},
			},
			{
				{Design: domain.CardDesignClover, Value: 7},
				{Design: domain.CardDesignClover, Value: 8},
				{Design: domain.CardDesignClover, Value: 9},
			},
		}, []int{2, 4})
	})

	t.Run("rearrange accepts the long form and rank letters", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)

		assert.Equal(t, mockOutput, c.Exec("rearrange sA,hA,dA / 0"))

		m.AssertCalled(t, "Play", [][]domain.MachiavelliCardRef{
			{
				{Design: domain.CardDesignSpade, Value: 1},
				{Design: domain.CardDesignHeart, Value: 1},
				{Design: domain.CardDesignDiamond, Value: 1},
			},
		}, []int{0})
	})

	t.Run("rearrange rejects malformed input without calling the interactor", func(t *testing.T) {
		// 手札を 1 枚も出さない組み替えはルール違反なので、ここで弾く。
		for _, arg := range []string{"ra", "ra s5,h5,d5", "ra / 1", "ra s5,x9 / 1", "ra s5,h5,d5 / x", "ra s5,h5,d5 /"} {
			m := newMock()
			c := controller.NewMachiavelliCuiController(m)

			out := c.Exec(arg)

			assert.Contains(t, out, i18n.T("usageRaRearrangeGroups"), "input %q", arg)
			m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
		}
	})
}
