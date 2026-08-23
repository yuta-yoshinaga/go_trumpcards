package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestRistikontraCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockRistikontraInteractor {
		m := new(mockUsecases.MockRistikontraInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultRistikontraConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewRistikontraCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("play command", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("play missing arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		out := c.Exec("p")
		assert.Contains(t, out, msgUsage("usagePHandidx"))
	})

	t.Run("play invalid arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		out := c.Exec("p xyz")
		assert.True(t, msgRejected(out))
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.RistikontraConfig) bool {
			return cfg.CpuDifficulty == domain.RistikontraDifficultyHard
		}))
	})

	t.Run("set difficulty invalid", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		out := c.Exec("sd 9")
		assert.Contains(t, out, msgInvalidCpuDifficultyPrefix())
	})

	// **席数を変えるコマンドは無い。** リスティコントラは常に 2 対 2 の
	// 固定パートナーシップなので、クローン元のピシュティにあった sp /
	// setplayers は消した。**宣伝も受理もしないこと**を両方見る。
	t.Run("set players is not a command", func(t *testing.T) {
		for _, cmd := range []string{"sp 3", "setplayers 3", "sp 4"} {
			m := newMock()
			c := controller.NewRistikontraCuiController(m)
			c.Exec(cmd)
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		}
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewRistikontraCuiController(m)
		out := c.Exec("zzz")
		assert.NotEmpty(t, out)
	})
}
