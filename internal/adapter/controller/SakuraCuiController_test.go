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

func TestSakuraCuiController_Exec(t *testing.T) {
	const mockOutput = `{"players":[]}`

	newMock := func() *mockUsecases.MockSakuraInteractor {
		m := new(mockUsecases.MockSakuraInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultSakuraConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewSakuraCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSakuraConfig())
	})
	t.Run("play hand only", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2, -1)
	})
	t.Run("play with a chosen field card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("p 2 3"))
		m.AssertCalled(t, "Play", 2, 3)
	})
	t.Run("play needs a card index", func(t *testing.T) {
		m := newMock()
		out := controller.NewSakuraCuiController(m).Exec("p")
		assert.Contains(t, out, "Card index is required")
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})
	t.Run("play rejects a non-numeric index", func(t *testing.T) {
		m := newMock()
		out := controller.NewSakuraCuiController(m).Exec("p x")
		assert.Contains(t, out, "Invalid card index")
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})
	// **場札インデックスが数字でなければ「指定なし」に倒す。** 途中で止めると
	// 手札は出せているのに手番が進まない。
	t.Run("play ignores a non-numeric field index", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("p 1 x"))
		m.AssertCalled(t, "Play", 1, -1)
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set seats", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("ss 4"))
		cfg := domain.DefaultSakuraConfig()
		cfg.Seats = 4
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set seats rejects out-of-range values", func(t *testing.T) {
		m := newMock()
		out := controller.NewSakuraCuiController(m).Exec("ss 9")
		assert.Contains(t, out, "Invalid number of seats")
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set rounds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSakuraCuiController(m).Exec("sr 5"))
		cfg := domain.DefaultSakuraConfig()
		cfg.Rounds = 5
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set rounds rejects zero", func(t *testing.T) {
		m := newMock()
		out := controller.NewSakuraCuiController(m).Exec("sr 0")
		assert.Contains(t, out, "Invalid number of rounds")
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewSakuraCuiController(m)
		assert.Equal(t, "hint", c.Exec("h"))
		assert.Equal(t, "log", c.Exec("l"))
	})
	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewSakuraCuiController(newMock()).Exec("pl 1")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "p")
	})
}

// 設定コマンドは範囲の両端を通す (境界で弾かれない)。
func TestSakuraCuiController_SettingBoundaries(t *testing.T) {
	tests := []struct {
		cmd  string
		want domain.SakuraConfig
	}{
		{"ss 2", domain.SakuraConfig{Seats: domain.SakuraMinSeats, Rounds: domain.SakuraDefaultRounds}},
		{"ss 4", domain.SakuraConfig{Seats: domain.SakuraMaxSeats, Rounds: domain.SakuraDefaultRounds}},
		{"sr 1", domain.SakuraConfig{Seats: domain.SakuraDefaultSeats, Rounds: domain.SakuraMinRounds}},
		{"sr 12", domain.SakuraConfig{Seats: domain.SakuraDefaultSeats, Rounds: domain.SakuraMaxRounds}},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			m := new(mockUsecases.MockSakuraInteractor)
			m.On("GetConfig").Return(domain.DefaultSakuraConfig())
			m.On("ResetWithConfig", mock.Anything).Return("ok")
			assert.Equal(t, "ok", controller.NewSakuraCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "ResetWithConfig", tt.want)
			// 送った設定はドメインの検証を通る。
			assert.NoError(t, tt.want.Validate())
		})
	}
}
