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

// **引き直しは Eight-Game Mix のためにある。** 同じコントローラーを 2 つの
// ゲームが共有しているので、番号の数え方も席数の可否もここが唯一の入口になる。
func TestHorseCuiController_Draw(t *testing.T) {
	const mockOutput = `{"seats":[]}`

	newMock := func(cfg domain.HorseConfig) *mockUsecases.MockHorseInteractor {
		m := new(mockUsecases.MockHorseInteractor)
		m.On("Exchange", mock.Anything).Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(cfg)
		return m
	}

	// **番号は 0 始まり。** 2-7 単体やムスと同じ数え方にしないと、同じ操作が
	// 1 枚ずれた札を捨てる。
	t.Run("draw sends zero-based indices", func(t *testing.T) {
		m := newMock(domain.DefaultEightGameConfig())
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("d 0 2"))
		m.AssertCalled(t, "Exchange", []int{0, 2})
	})

	t.Run("stand pat sends nothing", func(t *testing.T) {
		m := newMock(domain.DefaultEightGameConfig())
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("sp"))
		m.AssertCalled(t, "Exchange", []int(nil))
	})

	// **読めない番号があれば 1 枚も捨てない。** 残りを「別の合法な手」として
	// 打つと取り返しがつかない (#5390)。
	for _, cmd := range []string{"d", "d x", "d 9", "d 0 x"} {
		t.Run("rejects "+cmd, func(t *testing.T) {
			m := newMock(domain.DefaultEightGameConfig())
			out := controller.NewHorseCuiController(m).Exec(cmd)
			assert.NotEqual(t, mockOutput, out)
			m.AssertNotCalled(t, "Exchange", mock.Anything)
		})
	}

	// **席数の可否はバリアントで変わる。** Eight-Game Mix は 4 人卓だけ ──
	// 6 を通すと 6 種目目で理由も出さずにマッチが終わる。
	t.Run("eight-game rejects a six-seat table", func(t *testing.T) {
		m := newMock(domain.DefaultEightGameConfig())
		out := controller.NewHorseCuiController(m).Exec("ss 6")
		assert.Contains(t, out, msgKey("invalidNumberOfSeats469"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("horse still accepts a six-seat table", func(t *testing.T) {
		m := newMock(domain.DefaultHorseConfig())
		assert.Equal(t, mockOutput, controller.NewHorseCuiController(m).Exec("ss 6"))
		cfg := domain.DefaultHorseConfig()
		cfg.Seats = 6
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
}
