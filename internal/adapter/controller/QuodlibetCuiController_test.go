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

func TestQuodlibetCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	newMock := func() *mockUsecases.MockQuodlibetInteractor {
		m := new(mockUsecases.MockQuodlibetInteractor)
		m.On("GetConfig").Return(domain.DefaultQuodlibetConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("SelectContract", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextDeal").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewQuodlibetCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultQuodlibetConfig())
	})

	for _, cmd := range []string{"contract", "c"} {
		t.Run(cmd+" chooses a contract", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec(cmd+" 2"))
			m.AssertCalled(t, "SelectContract", 2)
		})
	}

	for _, cmd := range []string{"play", "p"} {
		t.Run(cmd+" plays a card", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec(cmd+" 3"))
			m.AssertCalled(t, "Play", 3)
		})
	}

	// **パスは -1 のプレイ。** 出せる札があるかの判定はドメインに 1 つだけ置く。
	t.Run("pass goes through play as -1", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Play", -1)
	})

	t.Run("contract and play need a number", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewQuodlibetCuiController(m).Exec("c"), msgStem("contractRequired"))
		assert.Contains(t, controller.NewQuodlibetCuiController(m).Exec("p"), msgStem("cardIndexRequired"))
	})

	// **範囲外の種目番号は弾く。** 12 以上はどの輪にも無い。
	t.Run("a contract outside 0-11 is refused", func(t *testing.T) {
		m := newMock()
		out := controller.NewQuodlibetCuiController(m).Exec("c 12")
		assert.Contains(t, out, msgStem("invalidContract"))
		m.AssertNotCalled(t, "SelectContract", 12)
	})

	for _, cmd := range []string{"nextdeal", "nd"} {
		t.Run(cmd+" advances the deal", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextDeal")
		})
	}

	t.Run("setdifficulty resets with the new difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultQuodlibetConfig()
		cfg.CpuDifficulty = domain.QuodlibetCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	// **auto は切り替え。** 同じ値を書き込むだけだと 12 回の選択が消えない。
	t.Run("auto toggles the automatic contract choice", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("auto"))
		cfg := domain.DefaultQuodlibetConfig()
		cfg.AutoSelectContract = !cfg.AutoSelectContract
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewQuodlibetCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewQuodlibetCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
