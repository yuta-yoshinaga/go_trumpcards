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

func mustSutdaOutputJSON(msg string) string {
	out := &controller.SutdaWebOutput{
		Players:       []*controller.SutdaWebOutputPlayer{},
		WinnerIdx:     -1,
		MaxRaises:     domain.SutdaMaxRaises,
		BetUnit:       domain.SutdaBetUnit,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSutdaOutputJSON: %v", err))
	}
	return string(b)
}

func TestSutdaWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"bet"}`

	diMock := new(usecase.MockSutdaInteractor)
	diMock.On("ResetWithConfig", domain.DefaultSutdaConfig()).Return(mockOutput)
	diMock.On("Action", mock.Anything).Return(mockOutput)
	diMock.On("NextHand").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewSutdaWebController(func() uc.SutdaInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.SutdaWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustSutdaOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	// **3 手は名前でも受ける。** action で包ませても得が無い。
	for _, cmd := range []string{"call", "raise", "fold"} {
		t.Run(cmd+" as its own command", func(t *testing.T) {
			run(t, fmt.Sprintf(`{"command":%q,"sessionId":"s1"}`, cmd), mockOutput, http.StatusOK)
			diMock.AssertCalled(t, "Action", cmd)
		})
	}
	t.Run("action carries the move", func(t *testing.T) {
		run(t, `{"command":"action","action":"raise","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Action", "raise")
	})
	t.Run("action without a move is refused", func(t *testing.T) {
		var input controller.SutdaWebInput
		_ = json.Unmarshal([]byte(`{"command":"action","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("nexthand", func(t *testing.T) {
		run(t, `{"command":"nexthand","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestSutdaWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name      string
		in        controller.SutdaWebConfig
		wantSeats int
		wantChips int
	}{
		{"defaults", controller.SutdaWebConfig{}, domain.SutdaDefaultSeats, domain.SutdaDefaultChips},
		{"in range", controller.SutdaWebConfig{Seats: i(5), StartChips: i(5000)}, 5, 5000},
		{"too many seats falls back", controller.SutdaWebConfig{Seats: i(99)}, domain.SutdaDefaultSeats, domain.SutdaDefaultChips},
		{"too few chips falls back", controller.SutdaWebConfig{StartChips: i(1)}, domain.SutdaDefaultSeats, domain.SutdaDefaultChips},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantSeats, cfg.Seats)
			assert.Equal(t, tt.wantChips, cfg.StartChips)
			assert.NoError(t, cfg.Validate())
		})
	}

	wc := controller.SutdaWebConfig{CpuDifficulty: i(9)}
	cfg := wc.ToConfig()
	assert.NoError(t, cfg.Validate(), "範囲外の難易度が丸められていない")
}

func TestSutdaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"bet"}`

	newMock := func() *usecase.MockSutdaInteractor {
		m := new(usecase.MockSutdaInteractor)
		m.On("GetConfig").Return(domain.DefaultSutdaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Action", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewSutdaCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSutdaConfig())
	})

	for _, tt := range []struct{ cmd, action string }{
		{"call", domain.SutdaActionCall}, {"c", domain.SutdaActionCall},
		{"raise", domain.SutdaActionRaise}, {"b", domain.SutdaActionRaise},
		{"fold", domain.SutdaActionFold}, {"f", domain.SutdaActionFold},
	} {
		t.Run(tt.cmd+" sends "+tt.action, func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "Action", tt.action)
		})
	}

	for _, cmd := range []string{"nexthand", "nh"} {
		t.Run(cmd+" advances the hand", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextHand")
		})
	}

	t.Run("setdifficulty and setseats reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultSutdaConfig()
		cfg.CpuDifficulty = domain.SutdaCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m2).Exec("ss 5"))
		cfg2 := domain.DefaultSutdaConfig()
		cfg2.Seats = 5
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **卓は 2〜5 席。** 1 席や 9 席は組めない。
	t.Run("a seat count outside 2-5 is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewSutdaCuiController(m).Exec("ss 1"), msgStem("invalidPlayerCount25"))
		assert.Contains(t, controller.NewSutdaCuiController(m).Exec("ss 9"), msgStem("invalidPlayerCount25"))
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewSutdaCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewSutdaCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
