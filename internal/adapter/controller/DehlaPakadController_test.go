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

func mustDehlaPakadOutputJSON(msg string) string {
	out := &controller.DehlaPakadWebOutput{
		Players:         []*controller.DehlaPakadWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrick:       []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		TeamTens:        []int{},
		TeamKots:        []int{},
		HandHistory:     []*controller.DehlaPakadWebOutputHand{},
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		PrevTrickWinner: -1,
		StreakTeam:      -1,
		WinnerTeam:      -1,
		HintTrumpSuit:   -1,
		TrickCount:      domain.DehlaPakadTrickCount,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustDehlaPakadOutputJSON: %v", err))
	}
	return string(b)
}

func TestDehlaPakadWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	diMock := new(usecase.MockDehlaPakadInteractor)
	diMock.On("ResetWithConfig", domain.DefaultDehlaPakadConfig()).Return(mockOutput)
	diMock.On("SelectTrump", 3).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("Play", 0).Return(mockOutput)
	diMock.On("NextHand").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewDehlaPakadWebController(func() uc.DehlaPakadInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.DehlaPakadWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustDehlaPakadOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("trump", func(t *testing.T) {
		run(t, `{"command":"trump","trumpSuit":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "SelectTrump", 3)
	})
	t.Run("trump without a suit is refused", func(t *testing.T) {
		var input controller.DehlaPakadWebInput
		_ = json.Unmarshal([]byte(`{"command":"trump","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 3)
	})
	// **札の指定が無い要求は断る。** 0 として通すと、意図しない札が出る。
	t.Run("play without a card index is refused", func(t *testing.T) {
		var input controller.DehlaPakadWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("card index 0 is a real card", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":0,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 0)
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
func TestDehlaPakadWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name     string
		in       controller.DehlaPakadWebConfig
		wantKots int
	}{
		{"defaults", controller.DehlaPakadWebConfig{}, domain.DehlaPakadDefaultKots},
		{"in range", controller.DehlaPakadWebConfig{TargetKots: i(5)}, 5},
		{"too many falls back", controller.DehlaPakadWebConfig{TargetKots: i(99)}, domain.DehlaPakadDefaultKots},
		{"zero falls back", controller.DehlaPakadWebConfig{TargetKots: i(0)}, domain.DehlaPakadDefaultKots},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantKots, cfg.TargetKots)
			assert.NoError(t, cfg.Validate())
		})
	}

	wc := controller.DehlaPakadWebConfig{CpuDifficulty: i(9)}
	cfg := wc.ToConfig()
	assert.NoError(t, cfg.Validate(), "範囲外の難易度が丸められていない")
}

func TestDehlaPakadCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	newMock := func() *usecase.MockDehlaPakadInteractor {
		m := new(usecase.MockDehlaPakadInteractor)
		m.On("GetConfig").Return(domain.DefaultDehlaPakadConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("SelectTrump", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewDehlaPakadCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultDehlaPakadConfig())
	})

	for _, cmd := range []string{"trump", "t"} {
		t.Run(cmd+" calls the trump", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec(cmd+" 3"))
			m.AssertCalled(t, "SelectTrump", 3)
		})
	}

	// **スートは 1-4 だけ。** 0 や 5 はどのスートでもない。
	t.Run("a suit outside 1-4 is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewDehlaPakadCuiController(m).Exec("t 0"), msgStem("invalidTrumpSuit"))
		assert.Contains(t, controller.NewDehlaPakadCuiController(m).Exec("t 5"), msgStem("invalidTrumpSuit"))
		m.AssertNotCalled(t, "SelectTrump", 0)
		m.AssertNotCalled(t, "SelectTrump", 5)
	})

	for _, cmd := range []string{"play", "p"} {
		t.Run(cmd+" plays a card", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec(cmd+" 3"))
			m.AssertCalled(t, "Play", 3)
		})
	}

	t.Run("trump and play need a number", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewDehlaPakadCuiController(m).Exec("t"), msgStem("suitRequired"))
		assert.Contains(t, controller.NewDehlaPakadCuiController(m).Exec("p"), msgStem("cardIndexRequired"))
	})

	for _, cmd := range []string{"nexthand", "nh"} {
		t.Run(cmd+" advances the hand", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextHand")
		})
	}

	t.Run("setdifficulty and setkots reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultDehlaPakadConfig()
		cfg.CpuDifficulty = domain.DehlaPakadCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m2).Exec("sk 5"))
		cfg2 := domain.DefaultDehlaPakadConfig()
		cfg2.TargetKots = 5
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewDehlaPakadCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewDehlaPakadCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
