//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

type scartoCuiTestPresenter struct{}

func (scartoCuiTestPresenter) Output(interfaces.ScartoGame, error) string   { return "" }
func (scartoCuiTestPresenter) ActionLogOutput(interfaces.ScartoGame) string { return "" }
func (scartoCuiTestPresenter) HintOutput(interfaces.ScartoGame) string      { return "" }

func TestScartoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":1}`

	newMock := func() *mockUsecases.MockScartoInteractor {
		m := new(mockUsecases.MockScartoInteractor)
		m.On("GetConfig").Return(domain.DefaultScartoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewScartoCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultScartoConfig())
	})

	t.Run("scarto three cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("scarto 0 1 2"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2})
	})

	t.Run("discard alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 3 4 5"))
		m.AssertCalled(t, "Discard", []int{3, 4, 5})
	})

	t.Run("scarto too few", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("scarto 0 1")
		assert.Contains(t, result, msgStem("threeIndicesRequiredScarto"))
	})

	t.Run("scarto invalid index", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("scarto 0 1 x")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultScartoConfig()
		expected.CpuDifficulty = domain.ScartoCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("settargetdeals validates, sets, and survives reset", func(t *testing.T) {
		interactor := usecase.NewScartoInteractor(domain.NewDefaultScarto(), scartoCuiTestPresenter{})
		c := controller.NewScartoCuiController(interactor)

		assert.Equal(t, "", c.Exec("settargetdeals 7"))
		assert.Equal(t, 7, interactor.GetConfig().TargetDeals)
		assert.Equal(t, "", c.Exec("reset"))
		assert.Equal(t, 7, interactor.GetConfig().TargetDeals)
	})

	t.Run("settargetdeals rejects invalid values", func(t *testing.T) {
		assert.Equal(t, "\x1eERR\x1e目標ディール数を指定してください (1-100)。", controller.NewScartoCuiController(newMock()).Exec("std"))
		assert.Equal(t, "\x1eERR\x1e無効な目標ディール数です: 0。1-100 を指定してください。", controller.NewScartoCuiController(newMock()).Exec("std 0"))
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
