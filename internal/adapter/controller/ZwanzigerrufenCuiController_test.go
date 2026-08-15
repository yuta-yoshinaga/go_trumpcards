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

func TestZwanzigerrufenCuiController_Exec(t *testing.T) {
	const mockOutput = `{"players":[]}`

	newMock := func() *mockUsecases.MockZwanzigerrufenInteractor {
		m := new(mockUsecases.MockZwanzigerrufenInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultZwanzigerrufenConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewZwanzigerrufenCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultZwanzigerrufenConfig())
	})
	t.Run("bid rufer", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("bid rufer"))
		m.AssertCalled(t, "Bid", domain.ZwanzigerrufenBidRufer)
	})
	t.Run("bid solo", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("bid solo"))
		m.AssertCalled(t, "Bid", domain.ZwanzigerrufenBidSolo)
	})
	// **trischaken は打てない。** 全員パスの結果としてしか成立しない契約なので、
	// コマンドで宣言できると「誰も落札しなかった」という前提が崩れる。
	t.Run("bid trischaken is rejected", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("bid trischaken")
		assert.Contains(t, out, "Invalid bid")
		m.AssertNotCalled(t, "Bid", mock.Anything)
	})
	t.Run("bid needs an argument", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("bid")
		assert.Contains(t, out, "Bid is required")
		m.AssertNotCalled(t, "Bid", mock.Anything)
	})
	t.Run("pass", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Pass")
	})
	t.Run("discard six indices", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput,
			controller.NewZwanzigerrufenCuiController(m).Exec("discard 0 1 2 3 4 5"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2, 3, 4, 5})
	})
	t.Run("discard needs six indices", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("discard 0 1 2")
		assert.Contains(t, out, "Six card indices")
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})
	t.Run("discard rejects a non-numeric index", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("discard 0 1 2 3 4 x")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})
	t.Run("play", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})
	t.Run("play rejects a non-numeric index", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("play x")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
		m.AssertNotCalled(t, "Play", mock.Anything)
	})
	t.Run("next trick and next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewZwanzigerrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("sd 2"))
		cfg := domain.DefaultZwanzigerrufenConfig()
		cfg.CpuDifficulty = domain.ZwanzigerrufenCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set difficulty rejects out-of-range", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("sd 9")
		assert.Contains(t, out, msgInvalidCpuDifficultyPrefix())
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set deals", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewZwanzigerrufenCuiController(m).Exec("st 6"))
		cfg := domain.DefaultZwanzigerrufenConfig()
		cfg.TargetDeals = 6
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set deals rejects out-of-range", func(t *testing.T) {
		m := newMock()
		out := controller.NewZwanzigerrufenCuiController(m).Exec("st 99")
		assert.Contains(t, out, "Invalid number of deals")
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewZwanzigerrufenCuiController(m)
		assert.Equal(t, "hint", c.Exec("h"))
		assert.Equal(t, "log", c.Exec("l"))
	})
	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewZwanzigerrufenCuiController(newMock()).Exec("bidd rufer")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "bid")
	})
}

// 設定コマンドが送る値はドメインの検証を通る。
func TestZwanzigerrufenCuiController_SettingsStayValid(t *testing.T) {
	tests := []struct {
		cmd  string
		want domain.ZwanzigerrufenConfig
	}{
		{"st 1", domain.ZwanzigerrufenConfig{
			CpuDifficulty: domain.ZwanzigerrufenCpuDifficultyNormal,
			TargetDeals:   domain.ZwanzigerrufenMinDeals,
		}},
		{"st 12", domain.ZwanzigerrufenConfig{
			CpuDifficulty: domain.ZwanzigerrufenCpuDifficultyNormal,
			TargetDeals:   domain.ZwanzigerrufenMaxDeals,
		}},
		{"sd 0", domain.ZwanzigerrufenConfig{
			CpuDifficulty: domain.ZwanzigerrufenCpuDifficultyEasy,
			TargetDeals:   domain.ZwanzigerrufenDefaultDeals,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			m := new(mockUsecases.MockZwanzigerrufenInteractor)
			m.On("GetConfig").Return(domain.DefaultZwanzigerrufenConfig())
			m.On("ResetWithConfig", mock.Anything).Return("ok")
			assert.Equal(t, "ok", controller.NewZwanzigerrufenCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "ResetWithConfig", tt.want)
			assert.NoError(t, tt.want.Validate())
		})
	}
}
