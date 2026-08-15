package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDaifugoCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`

	newMock := func() *mockUsecases.MockDaifugoInteractor {
		m := new(mockUsecases.MockDaifugoInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Sort", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultDaifugoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play command p with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("play command play with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("play")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("play command p with one index", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{2})
	})

	t.Run("play command play with multiple indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("play 0 3 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 3, 4})
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})

	t.Run("sort command default mode", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", mock.MatchedBy(func(mode interface{}) bool {
			return mode == domain.DaifugoSortByStrength
		}))
	})

	t.Run("sort command with mode argument", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", domain.DaifugoSortBySuit)
	})

	t.Run("sort command with invalid mode argument uses default", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort abc")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", domain.DaifugoSortByStrength)
	})

	t.Run("sort command with out-of-range mode uses default", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort 99")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", domain.DaifugoSortByStrength)
	})

	t.Run("play command ignores non-numeric index with warning", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p 0 abc 2")
		assert.Contains(t, result, "'abc'")
		assert.Contains(t, result, mockOutput)
		m.AssertCalled(t, "Play", []int{0, 2})
	})

	t.Run("play command no warning when all valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p 0 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 2})
	})
}

// --- setdifficulty ---

func TestDaifugoCuiController_SetDifficulty_Valid(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.CpuDifficulty = domain.DaifugoDifficultyHard
	mi.On("ResetWithConfig", cfg).Return("sd ok")
	assert.Equal(t, "sd ok", c.Exec("sd 2"))
}

func TestDaifugoCuiController_SetDifficulty_LongCommand(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.CpuDifficulty = domain.DaifugoDifficultyEasy
	mi.On("ResetWithConfig", cfg).Return("sd ok")
	assert.Equal(t, "sd ok", c.Exec("setdifficulty 1"))
}

func TestDaifugoCuiController_SetDifficulty_NoArgs(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequiredAlt())
}

func TestDaifugoCuiController_SetDifficulty_InvalidValue(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, msgInvalidCpuDifficulty("3"), c.Exec("sd 3"))
	assert.Equal(t, msgInvalidCpuDifficulty("-1"), c.Exec("sd -1"))
}

// --- setjoker ---

func TestDaifugoCuiController_SetJoker_Valid(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.JokerCount = 1
	mi.On("ResetWithConfig", cfg).Return("sj ok")
	assert.Equal(t, "sj ok", c.Exec("sj 1"))
}

func TestDaifugoCuiController_SetJoker_LongCommand(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.JokerCount = 0
	mi.On("ResetWithConfig", cfg).Return("sj ok")
	assert.Equal(t, "sj ok", c.Exec("setjoker 0"))
}

func TestDaifugoCuiController_SetJoker_NoArgs(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("sj"), "Joker count is required")
}

func TestDaifugoCuiController_SetJoker_InvalidValue(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("sj abc"), "Invalid joker count: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid joker count: 3. Please enter 0-2.", c.Exec("sj 3"))
	assert.Equal(t, "Invalid joker count: -1. Please enter 0-2.", c.Exec("sj -1"))
}

// --- setrule ---

func TestDaifugoCuiController_SetRule_Valid(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.FiveSkipEnabled = true
	mi.On("ResetWithConfig", cfg).Return("sr ok")
	assert.Equal(t, "sr ok", c.Exec("sr 5skip 1"))
}

func TestDaifugoCuiController_SetRule_LongCommand(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.EightCutEnabled = false
	mi.On("ResetWithConfig", cfg).Return("sr ok")
	assert.Equal(t, "sr ok", c.Exec("setrule 8cut 0"))
}

func TestDaifugoCuiController_SetRule_NoArgs(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	result := c.Exec("sr")
	assert.Contains(t, result, "Usage: sr <rule> <0|1> | sr list")
	assert.Contains(t, result, "8cut")
}

func TestDaifugoCuiController_SetRule_List(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	result := c.Exec("sr list")
	assert.Contains(t, result, "Rules:")
	// Default config: 8cut ON, 5skip OFF
	assert.Contains(t, result, "8cut")
	assert.Contains(t, result, "ON")
	assert.Contains(t, result, "5skip")
	assert.Contains(t, result, "OFF")
}

func TestDaifugoCuiController_SetRule_List_LongCommand(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	result := c.Exec("setrule list")
	assert.Contains(t, result, "Rules:")
}

func TestDaifugoCuiController_SetRule_OneArg(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("sr 8cut"), "Usage: sr <rule> <0|1> | sr list")
}

func TestDaifugoCuiController_SetRule_UnknownRule(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("sr unknown 1"), "Unknown rule: unknown")
}

func TestDaifugoCuiController_SetRule_InvalidValue(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("sr 8cut 2"), "Invalid value: 2")
	assert.Contains(t, c.Exec("sr 8cut abc"), "Invalid value: abc")
	assert.Contains(t, c.Exec("sr 8cut -1"), "Invalid value: -1")
}

func TestDaifugoCuiController_SetRule_AllKeys(t *testing.T) {
	tests := []struct {
		rule string
		val  string
		get  func(domain.DaifugoConfig) bool
	}{
		{"8cut", "0", func(c domain.DaifugoConfig) bool { return c.EightCutEnabled }},
		{"11back", "0", func(c domain.DaifugoConfig) bool { return c.ElevenBackEnabled }},
		{"seq", "0", func(c domain.DaifugoConfig) bool { return c.SequenceEnabled }},
		{"exchange", "0", func(c domain.DaifugoConfig) bool { return c.CardExchangeEnabled }},
		{"5skip", "1", func(c domain.DaifugoConfig) bool { return c.FiveSkipEnabled }},
		{"7pass", "1", func(c domain.DaifugoConfig) bool { return c.SevenPassEnabled }},
		{"10discard", "1", func(c domain.DaifugoConfig) bool { return c.TenDiscardEnabled }},
		{"spade3", "1", func(c domain.DaifugoConfig) bool { return c.SpadeThreeEnabled }},
		{"capital", "1", func(c domain.DaifugoConfig) bool { return c.CapitalFallEnabled }},
		{"9reverse", "1", func(c domain.DaifugoConfig) bool { return c.NineReverseEnabled }},
		{"coupdetat", "1", func(c domain.DaifugoConfig) bool { return c.CoupDetatEnabled }},
		{"numberlock", "1", func(c domain.DaifugoConfig) bool { return c.NumberLockEnabled }},
		{"sandstorm", "1", func(c domain.DaifugoConfig) bool { return c.SandstormEnabled }},
		{"emperor", "1", func(c domain.DaifugoConfig) bool { return c.EmperorEnabled }},
		{"seqrev", "1", func(c domain.DaifugoConfig) bool { return c.SequenceRevolutionEnabled }},
		{"seqlock", "1", func(c domain.DaifugoConfig) bool { return c.SequenceLockEnabled }},
		{"illegal", "1", func(c domain.DaifugoConfig) bool { return c.IllegalFinishEnabled }},
		{"12bomber", "1", func(c domain.DaifugoConfig) bool { return c.QueenBomberEnabled }},
		{"blindexchange", "1", func(c domain.DaifugoConfig) bool { return c.BlindExchangeEnabled }},
	}
	for _, tc := range tests {
		t.Run(tc.rule, func(t *testing.T) {
			mi := new(mockUsecases.MockDaifugoInteractor)
			c := controller.NewDaifugoCuiController(mi)
			mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
			expected := tc.val == "1"
			mi.On("ResetWithConfig", mock.MatchedBy(func(cfg domain.DaifugoConfig) bool {
				return tc.get(cfg) == expected
			})).Return("ok")
			assert.Equal(t, "ok", c.Exec("sr "+tc.rule+" "+tc.val))
		})
	}
}

// --- suitlockmode ---

func TestDaifugoCuiController_SuitLockMode_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.DaifugoSuitLockMode
	}{
		{"0", domain.DaifugoSuitLockNone},
		{"1", domain.DaifugoSuitLockPartial},
		{"2", domain.DaifugoSuitLockFull},
	}
	for _, tc := range tests {
		t.Run("suitlockmode "+tc.input, func(t *testing.T) {
			mi := new(mockUsecases.MockDaifugoInteractor)
			c := controller.NewDaifugoCuiController(mi)
			mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
			cfg := domain.DefaultDaifugoConfig()
			cfg.SuitLockMode = tc.expected
			mi.On("ResetWithConfig", cfg).Return("slm ok")
			assert.Equal(t, "slm ok", c.Exec("suitlockmode "+tc.input))
		})
	}
}

func TestDaifugoCuiController_SuitLockMode_NoArgs(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("suitlockmode"), "Suit lock mode is required")
}

func TestDaifugoCuiController_SuitLockMode_InvalidValue(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("suitlockmode abc"), "Invalid suit lock mode: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid suit lock mode: 3. Please enter 0-2.", c.Exec("suitlockmode 3"))
	assert.Equal(t, "Invalid suit lock mode: -1. Please enter 0-2.", c.Exec("suitlockmode -1"))
}

// --- 5skipcount ---

func TestDaifugoCuiController_FiveSkipCount_Valid(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDaifugoConfig())
	cfg := domain.DefaultDaifugoConfig()
	cfg.FiveSkipCount = 3
	mi.On("ResetWithConfig", cfg).Return("5sc ok")
	assert.Equal(t, "5sc ok", c.Exec("5skipcount 3"))
}

func TestDaifugoCuiController_FiveSkipCount_NoArgs(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	assert.Contains(t, c.Exec("5skipcount"), "Five skip count is required")
}

func TestDaifugoCuiController_FiveSkipCount_InvalidValue(t *testing.T) {
	mi := new(mockUsecases.MockDaifugoInteractor)
	c := controller.NewDaifugoCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("5skipcount abc"), "Invalid five skip count: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid five skip count: 0. Please enter 1-5.", c.Exec("5skipcount 0"))
	assert.Equal(t, "Invalid five skip count: 6. Please enter 1-5.", c.Exec("5skipcount 6"))
}
