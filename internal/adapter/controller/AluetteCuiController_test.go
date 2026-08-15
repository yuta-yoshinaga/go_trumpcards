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

func TestAluetteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockAluetteInteractor {
		m := new(mockUsecases.MockAluetteInteractor)
		m.On("GetConfig").Return(domain.DefaultAluetteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewAluetteCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewAluetteCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAluetteCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultAluetteConfig())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAluetteCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewAluetteCuiController(newMock()).Exec("play"), msgCardIndexRequired())
	})

	t.Run("play with a non-numeric index", func(t *testing.T) {
		assert.Contains(t, controller.NewAluetteCuiController(newMock()).Exec("play x"), msgInvalidCardIndexPrefix())
	})

	// **捨て札も入札もこのゲームには無い。**タロー系から写すと scarto / bid が
	// 付いてくるが、黙って別の動作に落ちてはならない。
	t.Run("scarto and bidding commands are not accepted", func(t *testing.T) {
		for _, cmd := range []string{"scarto 0 1", "discard 0 1", "bid 1", "pass"} {
			assert.Contains(t, controller.NewAluetteCuiController(newMock()).Exec(cmd), "コマンドが不明です")
		}
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewAluetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAluetteCuiController(m).Exec("sd 2"))
		expected := domain.DefaultAluetteConfig()
		expected.CpuDifficulty = domain.AluetteCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewAluetteCuiController(newMock()).Exec("sd 9"), "Invalid CPU difficulty")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewAluetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewAluetteCuiController(newMock()).Exec("zzz"), "コマンドが不明です")
	})
}
