//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCoincheCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCoincheInteractor {
		m := new(mockUsecases.MockCoincheInteractor)
		m.On("GetConfig").Return(domain.DefaultCoincheConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Coinche").Return(mockOutput)
		m.On("Surcoinche").Return(mockOutput)
		m.On("DeclineDouble").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit long", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
	})

	// **点とスートは 2 つで 1 つの宣言。** 片方でも落ちると、宣言した契約と
	// 盤面の切り札が食い違う。
	t.Run("bid carries both the points and the suit", func(t *testing.T) {
		for _, input := range []string{"b 110 1", "bid 110 1"} {
			m := newMock()
			c := controller.NewCoincheCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(input))
			m.AssertCalled(t, "Bid", 110, 1)
		}
	})

	// 片方だけ、値が範囲外、数字でない — どれも宣言として通してはいけない。
	t.Run("bid rejects an incomplete or out-of-range declaration", func(t *testing.T) {
		// 85 は契約表に無い値だが、表を知っているのはドメイン側。ここは
		// 範囲と欠落だけを見る。
		for _, input := range []string{"b", "b 110", "b 110 0", "b 110 9", "b xx 1", "b 110 yy", "b 70 1", "b 300 1"} {
			m := newMock()
			c := controller.NewCoincheCuiController(m)
			c.Exec(input)
			m.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything)
		}
	})

	t.Run("pass pa", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pa"))
		m.AssertCalled(t, "Pass")
	})

	t.Run("doubling commands", func(t *testing.T) {
		for _, tc := range []struct{ input, method string }{
			{"co", "Coinche"}, {"coinche", "Coinche"},
			{"su", "Surcoinche"}, {"surcoinche", "Surcoinche"},
			{"ok", "DeclineDouble"},
		} {
			m := newMock()
			c := controller.NewCoincheCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(tc.input))
			m.AssertCalled(t, tc.method)
		}
	})

	t.Run("bid names the offending suit", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("b 110 abc"), msgStem("invalidSuit"))
		assert.Contains(t, c.Exec("b 110 5"), msgKey("invalidSuit", "val", "5"))
	})

	t.Run("play p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid arg", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultCoincheConfig()
		expected.CpuDifficulty = domain.CoincheCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
	})

	t.Run("settarget valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 500"))
		expected := domain.DefaultCoincheConfig()
		expected.TargetScore = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget no args", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("st"), msgTargetScoreRequired())
	})

	t.Run("settarget invalid", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("st 0"), msgInvalidTargetScore("0"))
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewCoincheCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewCoincheCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
