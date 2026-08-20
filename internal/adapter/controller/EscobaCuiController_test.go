package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestEscobaCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockEscobaInteractor {
		m := new(mockUsecases.MockEscobaInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultEscobaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewEscobaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("play with capture", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1 2"))
		m.AssertCalled(t, "Play", 0, []int{1, 2})
	})

	t.Run("play lay (no table)", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0"))
		m.AssertCalled(t, "Play", 0, []int{})
	})

	t.Run("play missing hand", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Contains(t, c.Exec("p"), msgUsage("usagePHandidxTableidxMany"))
	})

	t.Run("play bad hand index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.True(t, msgRejected(c.Exec("p xyz")))
	})

	t.Run("sd (difficulty)", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.EscobaConfig) bool {
			return cfg.CpuDifficulty == domain.EscobaCpuDifficultyHard
		}))
	})

	t.Run("st (target)", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 21"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.EscobaConfig) bool {
			return cfg.TargetScore == 21
		}))
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEscobaCuiController(m)
		assert.Contains(t, c.Exec("p 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
