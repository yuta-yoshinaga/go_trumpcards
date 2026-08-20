package controller_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestScopaCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockScopaInteractor {
		m := new(mockUsecases.MockScopaInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultScopaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewScopaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("play with capture", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1 2"))
		m.AssertCalled(t, "Play", 0, []int{1, 2})
	})

	t.Run("play lay (no table)", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0"))
		m.AssertCalled(t, "Play", 0, []int{})
	})

	t.Run("play missing hand", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Contains(t, c.Exec("p"), msgUsage("usagePHandidxTableidxMany"))
	})

	t.Run("play bad hand index", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.True(t, msgRejected(c.Exec("p xyz")))
	})

	t.Run("sd (difficulty)", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ScopaConfig) bool {
			return cfg.CpuDifficulty == domain.ScopaDifficultyHard
		}))
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})
}

// #5619: Web の設定パネルは目標点 (11/16/21) を変えられるのに、CUI には
// `sd`/`setdifficulty` しか無く、目標点は既定の 11 に固定されていた。
func TestScopaCuiControllerSetsTheTargetScore(t *testing.T) {
	// **既定と違う設定を土台にする。**既定のままだと「今の設定を読む」実装と
	// 「既定から作り直す」実装を区別できない。
	current := domain.DefaultScopaConfig()
	current.CpuDifficulty = domain.ScopaDifficultyHard
	newMock := func() *mockUsecases.MockScopaInteractor {
		m := new(mockUsecases.MockScopaInteractor)
		m.On("GetConfig").Return(current)
		m.On("ResetWithConfig", mock.Anything).Return("reset-ok")
		return m
	}

	for _, alias := range []string{"st", "settarget"} {
		t.Run(alias, func(t *testing.T) {
			m := newMock()
			c := controller.NewScopaCuiController(m)

			assert.Equal(t, "reset-ok", c.Exec(alias+" 21"))
			m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ScopaConfig) bool {
				// 目標点だけ変わり、難易度は土台のまま。
				return cfg.TargetScore == 21 && cfg.CpuDifficulty == domain.ScopaDifficultyHard
			}))
		})
	}

	// 範囲はドメインの Validate と同じ。外れた値でリセットしない。
	for _, v := range []int{domain.ScopaMinTargetScore - 1, domain.ScopaMaxTargetScore + 1} {
		t.Run("rejects "+strconv.Itoa(v), func(t *testing.T) {
			m := newMock()
			out := controller.NewScopaCuiController(m).Exec("st " + strconv.Itoa(v))
			assert.Contains(t, out, strconv.Itoa(domain.ScopaMaxTargetScore))
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		})
	}

	t.Run("asks for the value when it is missing", func(t *testing.T) {
		m := newMock()
		assert.NotEmpty(t, controller.NewScopaCuiController(m).Exec("st"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
}
