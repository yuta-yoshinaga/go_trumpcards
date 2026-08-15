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

func TestTarocchiniCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTarocchiniInteractor {
		m := new(mockUsecases.MockTarocchiniInteractor)
		m.On("GetConfig").Return(domain.DefaultTarocchiniConfig())
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
		assert.Equal(t, "bye.", controller.NewTarocchiniCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewTarocchiniCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarocchiniCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTarocchiniConfig())
	})

	t.Run("scarto and its discard alias", func(t *testing.T) {
		for _, cmd := range []string{"scarto 0 3", "discard 0 3"} {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewTarocchiniCuiController(m).Exec(cmd))
			m.AssertCalled(t, "Discard", []int{0, 3})
		}
	})

	// 案内文の枚数は TarocchiniSurplus から出す。数字を直接書くと余剰枚数が
	// 変わったときに案内だけ古くなる。
	t.Run("scarto with too few indices names the required count", func(t *testing.T) {
		out := controller.NewTarocchiniCuiController(newMock()).Exec("scarto 0")
		assert.Contains(t, out, "2 card indices are required")
	})

	t.Run("scarto with a non-numeric index", func(t *testing.T) {
		out := controller.NewTarocchiniCuiController(newMock()).Exec("scarto x y")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarocchiniCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewTarocchiniCuiController(newMock()).Exec("play"), msgCardIndexRequired())
	})

	// このゲームに入札は無い。bid/pass が黙って別の動作に落ちてはならない。
	t.Run("bidding commands are not accepted", func(t *testing.T) {
		for _, cmd := range []string{"bid 1", "pass"} {
			assert.Contains(t, controller.NewTarocchiniCuiController(newMock()).Exec(cmd), "コマンドが不明です")
		}
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTarocchiniCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarocchiniCuiController(m).Exec("sd 2"))
		expected := domain.DefaultTarocchiniConfig()
		expected.CpuDifficulty = domain.TarocchiniCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewTarocchiniCuiController(newMock()).Exec("sd 9"), "Invalid CPU difficulty")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTarocchiniCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewTarocchiniCuiController(newMock()).Exec("zzz"), "コマンドが不明です")
	})
}
