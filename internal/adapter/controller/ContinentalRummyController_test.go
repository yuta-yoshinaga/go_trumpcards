//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustContinentalRummyOutputJSON(msg string) string {
	out := &controller.ContinentalRummyWebOutput{
		Players:        []*controller.ContinentalRummyWebOutputPlayer{},
		Layouts:        domain.ContinentalRummyLayouts(),
		WinnerIdx:      -1,
		GoOutIdx:       -1,
		HintDiscardIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustContinentalRummyOutputJSON: %v", err))
	}
	return string(b)
}

func newContinentalRummyMock(mockOutput string) *usecase.MockContinentalRummyInteractor {
	m := new(usecase.MockContinentalRummyInteractor)
	m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	m.On("DrawStock").Return(mockOutput)
	m.On("DrawDiscard").Return(mockOutput)
	m.On("Discard", mock.Anything).Return(mockOutput)
	m.On("GoOut", mock.Anything).Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)
	m.On("GetConfig").Return(domain.DefaultContinentalRummyConfig())
	return m
}

func TestContinentalRummyWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"draw"}`
	diMock := newContinentalRummyMock(mockOutput)

	ctrl := controller.NewContinentalRummyWebController(func() uc.ContinentalRummyInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.ContinentalRummyWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustContinentalRummyOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultContinentalRummyConfig())
	})

	// **山と捨て札は別のコマンド。** 真偽値に畳むと、付け忘れた要求が
	// 黙ってどちらかに倒れる。
	for _, tc := range []struct{ cmd, method, other string }{
		{"stock", "DrawStock", "DrawDiscard"}, {"ds", "DrawStock", "DrawDiscard"},
		{"take", "DrawDiscard", "DrawStock"}, {"dd", "DrawDiscard", "DrawStock"},
	} {
		t.Run(tc.cmd+" calls "+tc.method, func(t *testing.T) {
			m := newContinentalRummyMock(mockOutput)
			c := controller.NewContinentalRummyWebController(func() uc.ContinentalRummyInteractorIF { return m })
			defer c.Stop()
			var input controller.ContinentalRummyWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+tc.cmd+`","sessionId":"s1"}`), &input)
			execRequest(t, c.Exec, &input).BodyIs(mockOutput)
			m.AssertCalled(t, tc.method)
			m.AssertNotCalled(t, tc.other)
		})
	}

	t.Run("discard and goout carry the hand index", func(t *testing.T) {
		run(t, `{"command":"discard","handIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Discard", 3)
		run(t, `{"command":"goout","handIndex":7,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "GoOut", 7)
	})

	// **上がるときも捨てる 1 枚を名指す。** 既定に落とすと、意図しない札が飛ぶ。
	t.Run("discard and goout without a hand index are refused", func(t *testing.T) {
		m := newContinentalRummyMock(mockOutput)
		c := controller.NewContinentalRummyWebController(func() uc.ContinentalRummyInteractorIF { return m })
		defer c.Stop()
		for _, cmd := range []string{"discard", "goout"} {
			var input controller.ContinentalRummyWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"s1"}`), &input)
			recorded := execRequest(t, c.Exec, &input)
			// **足りない引数は 400 で断る。** 既定に落として動かさない。
			recorded.CodeIs(http.StatusBadRequest)
			assert.Contains(t, recorded.Body.String(), "handIndex")
		}
		m.AssertNotCalled(t, "Discard", mock.Anything)
		m.AssertNotCalled(t, "GoOut", mock.Anything)
	})

	t.Run("next, hint and log", func(t *testing.T) {
		run(t, `{"command":"next","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "NextRound")
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})

	// **範囲外は既定に戻す (webutil.BoundedIntPtr の約束)。**
	t.Run("out-of-range config falls back to the default, in range it is kept", func(t *testing.T) {
		var bad controller.ContinentalRummyWebInput
		_ = json.Unmarshal([]byte(
			`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":9,"totalRounds":99}}`), &bad)
		assert.Equal(t, domain.DefaultContinentalRummyConfig(), bad.ToConfig())

		var good controller.ContinentalRummyWebInput
		_ = json.Unmarshal([]byte(
			`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":0,"totalRounds":10}}`), &good)
		cfg := good.ToConfig()
		assert.Equal(t, domain.ContinentalRummyCpuDifficultyEasy, cfg.CpuDifficulty)
		assert.Equal(t, domain.ContinentalRummyMaxRounds, cfg.TotalRounds)
		assert.NotEqual(t, domain.DefaultContinentalRummyConfig(), cfg)
	})
}

func TestContinentalRummyCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"draw"}`

	for _, tc := range []struct{ cmd, method string }{
		{"stock", "DrawStock"}, {"ds", "DrawStock"},
		{"take", "DrawDiscard"}, {"dd", "DrawDiscard"},
		{"next", "NextRound"}, {"n", "NextRound"},
	} {
		t.Run(tc.cmd+" calls "+tc.method, func(t *testing.T) {
			m := newContinentalRummyMock(mockOutput)
			assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m).Exec(tc.cmd))
			m.AssertCalled(t, tc.method)
		})
	}

	for _, tc := range []struct{ cmd, method string }{
		{"discard 2", "Discard"}, {"d 2", "Discard"}, {"goout 2", "GoOut"}, {"g 2", "GoOut"},
	} {
		t.Run(tc.cmd+" calls "+tc.method+" with the index", func(t *testing.T) {
			m := newContinentalRummyMock(mockOutput)
			assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m).Exec(tc.cmd))
			m.AssertCalled(t, tc.method, 2)
		})
	}

	t.Run("discard and goout without an index are refused", func(t *testing.T) {
		m := newContinentalRummyMock(mockOutput)
		for _, cmd := range []string{"discard", "goout"} {
			assert.Contains(t, controller.NewContinentalRummyCuiController(m).Exec(cmd),
				msgStem("cardIndexRequired"))
			assert.Contains(t, controller.NewContinentalRummyCuiController(m).Exec(cmd+" zz"),
				msgStem("invalidCardIndex"))
		}
		m.AssertNotCalled(t, "Discard", mock.Anything)
		m.AssertNotCalled(t, "GoOut", mock.Anything)
	})

	t.Run("sd and sr reset with the new setting", func(t *testing.T) {
		m := newContinentalRummyMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultContinentalRummyConfig()
		cfg.CpuDifficulty = domain.ContinentalRummyCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newContinentalRummyMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m2).Exec("sr 10"))
		cfg2 := domain.DefaultContinentalRummyConfig()
		cfg2.TotalRounds = domain.ContinentalRummyMaxRounds
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **範囲外は断り、範囲を名指す。**
	t.Run("an out-of-range round count is refused and names the range", func(t *testing.T) {
		m := newContinentalRummyMock(mockOutput)
		out := controller.NewContinentalRummyCuiController(m).Exec("sr 99")
		assert.Contains(t, out, msgStem("continentalrummy.invalidRounds"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.ContinentalRummyMaxRounds))
		assert.Contains(t, controller.NewContinentalRummyCuiController(m).Exec("sr 0"),
			msgStem("continentalrummy.invalidRounds"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newContinentalRummyMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewContinentalRummyCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewContinentalRummyCuiController(newContinentalRummyMock(mockOutput)).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
