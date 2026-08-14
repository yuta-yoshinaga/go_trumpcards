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

func TestTrogguCuiController_Exec(t *testing.T) {
	const mockOutput = `{"players":[]}`

	newMock := func() *mockUsecases.MockTrogguInteractor {
		m := new(mockUsecases.MockTrogguInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultTrogguConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTrogguCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTrogguConfig())
	})

	// **契約は 4 つとも打てる。** どれか一つでも綴りが違うと、その契約だけが
	// 静かに遊べなくなる。
	for _, tt := range []struct {
		arg  string
		want domain.TrogguBid
	}{
		{"trois", domain.TrogguBidTrois},
		{"solo", domain.TrogguBidSolo},
		{"piccolo", domain.TrogguBidPiccolo},
		{"misere", domain.TrogguBidMisere},
	} {
		t.Run("bid "+tt.arg, func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("bid "+tt.arg))
			m.AssertCalled(t, "Bid", tt.want)
		})
	}

	t.Run("bid rejects an unknown contract", func(t *testing.T) {
		m := newMock()
		out := controller.NewTrogguCuiController(m).Exec("bid nonsense")
		assert.Contains(t, out, "Invalid bid")
		m.AssertNotCalled(t, "Bid", mock.Anything)
	})
	t.Run("bid needs an argument", func(t *testing.T) {
		m := newMock()
		out := controller.NewTrogguCuiController(m).Exec("bid")
		assert.Contains(t, out, "Bid is required")
		m.AssertNotCalled(t, "Bid", mock.Anything)
	})
	t.Run("pass", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Pass")
	})
	t.Run("play", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("play 5"))
		m.AssertCalled(t, "Play", 5)
	})
	t.Run("play rejects a non-numeric index", func(t *testing.T) {
		m := newMock()
		out := controller.NewTrogguCuiController(m).Exec("play x")
		assert.Contains(t, out, "Invalid card index")
		m.AssertNotCalled(t, "Play", mock.Anything)
	})
	t.Run("next trick and next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewTrogguCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("sd 2"))
		cfg := domain.DefaultTrogguConfig()
		cfg.CpuDifficulty = domain.TrogguCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set difficulty rejects out-of-range", func(t *testing.T) {
		m := newMock()
		out := controller.NewTrogguCuiController(m).Exec("sd 9")
		assert.Contains(t, out, "Invalid CPU difficulty")
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set deals", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrogguCuiController(m).Exec("st 8"))
		cfg := domain.DefaultTrogguConfig()
		cfg.TargetDeals = 8
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set deals rejects out-of-range", func(t *testing.T) {
		m := newMock()
		out := controller.NewTrogguCuiController(m).Exec("st 99")
		assert.Contains(t, out, "Invalid number of deals")
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTrogguCuiController(m)
		assert.Equal(t, "hint", c.Exec("h"))
		assert.Equal(t, "log", c.Exec("l"))
	})
	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewTrogguCuiController(newMock()).Exec("bidd solo")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "bid")
	})
}

// 設定コマンドが送る値はドメインの検証を通る。
func TestTrogguCuiController_SettingsStayValid(t *testing.T) {
	tests := []struct {
		cmd  string
		want domain.TrogguConfig
	}{
		{"st 1", domain.TrogguConfig{
			CpuDifficulty: domain.TrogguCpuDifficultyNormal, TargetDeals: domain.TrogguMinDeals,
		}},
		{"st 12", domain.TrogguConfig{
			CpuDifficulty: domain.TrogguCpuDifficultyNormal, TargetDeals: domain.TrogguMaxDeals,
		}},
		{"sd 0", domain.TrogguConfig{
			CpuDifficulty: domain.TrogguCpuDifficultyEasy, TargetDeals: domain.TrogguDefaultDeals,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			m := new(mockUsecases.MockTrogguInteractor)
			m.On("GetConfig").Return(domain.DefaultTrogguConfig())
			m.On("ResetWithConfig", mock.Anything).Return("ok")
			assert.Equal(t, "ok", controller.NewTrogguCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "ResetWithConfig", tt.want)
			assert.NoError(t, tt.want.Validate())
		})
	}
}
