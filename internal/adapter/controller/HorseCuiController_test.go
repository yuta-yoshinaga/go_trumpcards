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

func TestHorseCuiController_Exec(t *testing.T) {
	const mockOutput = `{"seats":[]}`

	newMock := func() *mockUsecases.MockHorseInteractor {
		m := new(mockUsecases.MockHorseInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Action", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultHorseConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewHorseCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultHorseConfig())
	})

	// **ベッティングの語彙は種目に合わせる。** 1 つでも綴りが違うと、その手だけが
	// 打てなくなる。
	for _, tt := range []struct {
		cmd    string
		action int
		amount int
	}{
		{"f", domain.HoldemActionFold, 0},
		{"fold", domain.HoldemActionFold, 0},
		{"x", domain.HoldemActionCheck, 0},
		{"c", domain.HoldemActionCall, 0},
		{"b 50", domain.HoldemActionBet, 50},
		{"raise 120", domain.HoldemActionRaise, 120},
		{"allin", domain.HoldemActionAllIn, 0},
	} {
		t.Run(tt.cmd, func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "Action", tt.action, tt.amount, 0)
		})
	}

	t.Run("bet needs an amount", func(t *testing.T) {
		m := newMock()
		out := controller.NewHorseCuiController(m).Exec("b")
		assert.Contains(t, out, msgStem("betAmountRequired"))
		m.AssertNotCalled(t, "Action", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("bet rejects a non-numeric amount", func(t *testing.T) {
		m := newMock()
		out := controller.NewHorseCuiController(m).Exec("b x")
		assert.Contains(t, out, msgStem("invalidAmount"))
		m.AssertNotCalled(t, "Action", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("next hand", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("n"))
		m.AssertCalled(t, "NextHand")
	})

	// **席数は 4 / 6 / 9 だけ。** 種目側の卓サイズと同じものしか作れない。
	for _, seats := range []int{4, 6, 9} {
		t.Run("set seats "+string(rune('0'+seats)), func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput,
				controller.NewHorseCuiController(m).Exec("ss "+string(rune('0'+seats))))
			cfg := domain.DefaultHorseConfig()
			cfg.Seats = seats
			m.AssertCalled(t, "ResetWithConfig", cfg)
		})
	}
	// **引数なしの経路は別物。** WithParsedIntKeys の missing 分岐ではなく手前の
	// ガードが返すので、鍵を足しただけでは到達せず英語のまま残る。
	t.Run("set seats with no argument asks for one", func(t *testing.T) {
		m := newMock()
		out := controller.NewHorseCuiController(m).Exec("ss")
		assert.Equal(t, msgKey("numberOfSeatsRequired469"), out)
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	for _, seats := range []string{"3", "5", "8", "10"} {
		t.Run("set seats rejects "+seats, func(t *testing.T) {
			m := newMock()
			out := controller.NewHorseCuiController(m).Exec("ss " + seats)
			assert.Contains(t, out, msgKey("invalidNumberOfSeats469"))
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		})
	}
	t.Run("set hands", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("sh 5"))
		cfg := domain.DefaultHorseConfig()
		cfg.HandsPerDiscipline = 5
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
	t.Run("set hands rejects out-of-range", func(t *testing.T) {
		m := newMock()
		out := controller.NewHorseCuiController(m).Exec("sh 99")
		assert.Contains(t, out, msgStem("invalidNumberOfHands110"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewHorseCuiController(m)
		assert.Equal(t, "hint", c.Exec("h"))
		assert.Equal(t, "log", c.Exec("l"))
	})
	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewHorseCuiController(newMock()).Exec("foldd")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "fold")
	})
}

// 設定コマンドが送る値はドメインの検証を通る。
func TestHorseCuiController_SettingsStayValid(t *testing.T) {
	tests := []struct {
		cmd  string
		want domain.HorseConfig
	}{
		{"ss 9", domain.HorseConfig{
			Seats: 9, InitialChips: domain.HorseDefaultChips,
			HandsPerDiscipline: domain.HorseDefaultHandsPerDiscipline,
		}},
		{"sh 1", domain.HorseConfig{
			Seats: domain.HorseDefaultSeats, InitialChips: domain.HorseDefaultChips,
			HandsPerDiscipline: domain.HorseMinHandsPerDiscipline,
		}},
		{"sh 10", domain.HorseConfig{
			Seats: domain.HorseDefaultSeats, InitialChips: domain.HorseDefaultChips,
			HandsPerDiscipline: domain.HorseMaxHandsPerDiscipline,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			m := new(mockUsecases.MockHorseInteractor)
			m.On("GetConfig").Return(domain.DefaultHorseConfig())
			m.On("ResetWithConfig", mock.Anything).Return("ok")
			assert.Equal(t, "ok", controller.NewHorseCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "ResetWithConfig", tt.want)
			assert.NoError(t, tt.want.Validate())
		})
	}
}
