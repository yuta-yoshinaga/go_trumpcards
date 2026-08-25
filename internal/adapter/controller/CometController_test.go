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

func mustCometOutputJSON(msg string) string {
	out := &controller.CometWebOutput{
		Players:       []*controller.CometWebOutputPlayer{},
		Pile:          []*controller.WebOutputCard{},
		PlayableIdxs:  []int{},
		LastPlayerIdx: -1,
		WinnerIdx:     -1,
		HintHandIdx:   -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCometOutputJSON: %v", err))
	}
	return string(b)
}

func TestCometWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	diMock := new(usecase.MockCometInteractor)
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("Play", mock.Anything).Return(mockOutput)
	diMock.On("Pass").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewCometWebController(func() uc.CometInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.CometWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustCometOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultCometConfig())
	})
	t.Run("play carries the hand index", func(t *testing.T) {
		run(t, `{"command":"play","handIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 3)
	})
	// **0 は「省略」ではない。** handIndex がポインタでないと、札を指定しない
	// 要求が先頭の札を出す手として通ってしまう。
	t.Run("play without a hand index is refused", func(t *testing.T) {
		var input controller.CometWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Pass")
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "NextRound")
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestCometWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name       string
		in         controller.CometWebConfig
		wantSeats  int
		wantTarget int
	}{
		{"defaults", controller.CometWebConfig{},
			domain.CometDefaultPlayers, domain.CometDefaultTarget},
		{"in range", controller.CometWebConfig{Players: i(3), TargetScore: i(50)}, 3, 50},
		{"too many seats falls back", controller.CometWebConfig{Players: i(9)},
			domain.CometDefaultPlayers, domain.CometDefaultTarget},
		{"too few seats falls back", controller.CometWebConfig{Players: i(1)},
			domain.CometDefaultPlayers, domain.CometDefaultTarget},
		{"too high a target falls back", controller.CometWebConfig{TargetScore: i(999)},
			domain.CometDefaultPlayers, domain.CometDefaultTarget},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantSeats, cfg.Players)
			assert.Equal(t, tt.wantTarget, cfg.TargetScore)
			assert.NoError(t, cfg.Validate())
		})
	}
	wc := controller.CometWebConfig{CpuDifficulty: i(9)}
	assert.NoError(t, wc.ToConfig().Validate(), "範囲外の難易度が丸められていない")
	assert.Equal(t, domain.DefaultCometConfig(), controller.CometWebInput{}.ToConfig())
}

func TestCometCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	newMock := func() *usecase.MockCometInteractor {
		m := new(usecase.MockCometInteractor)
		m.On("GetConfig").Return(domain.DefaultCometConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCometCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCometConfig())
	})

	for _, cmd := range []string{"p 2", "play 2"} {
		t.Run(cmd+" plays a card", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec(cmd))
			m.AssertCalled(t, "Play", 2)
		})
	}

	t.Run("play without a hand index is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewCometCuiController(m).Exec("p"), msgStem("cardIndexRequired"))
		assert.Contains(t, controller.NewCometCuiController(m).Exec("p zz"), msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Play", mock.Anything)
	})

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Pass")
	})

	for _, cmd := range []string{"nextround", "nr"} {
		t.Run(cmd+" advances the round", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextRound")
		})
	}

	t.Run("setdifficulty, setplayers and settarget reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultCometConfig()
		cfg.CpuDifficulty = domain.CometCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m2).Exec("sp 3"))
		cfg2 := domain.DefaultCometConfig()
		cfg2.Players = 3
		m2.AssertCalled(t, "ResetWithConfig", cfg2)

		m3 := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m3).Exec("st 50"))
		cfg3 := domain.DefaultCometConfig()
		cfg3.TargetScore = 50
		m3.AssertCalled(t, "ResetWithConfig", cfg3)
	})

	// **卓は 2〜5 席。** 断る文言も範囲を名指す。
	t.Run("a seat count outside 2-5 is refused and names the range", func(t *testing.T) {
		m := newMock()
		out := controller.NewCometCuiController(m).Exec("sp 1")
		assert.Contains(t, out, msgStem("comet.invalidPlayerCount"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CometMinPlayers))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CometMaxPlayers))
		assert.Contains(t, controller.NewCometCuiController(m).Exec("sp 9"),
			msgStem("comet.invalidPlayerCount"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	// **勝負点は 20〜200。**
	t.Run("a target outside 20-200 is refused and names the range", func(t *testing.T) {
		m := newMock()
		out := controller.NewCometCuiController(m).Exec("st 5")
		assert.Contains(t, out, msgStem("comet.invalidTargetScore"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CometMinTarget))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CometMaxTarget))
		assert.Contains(t, controller.NewCometCuiController(m).Exec("st 999"),
			msgStem("comet.invalidTargetScore"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewCometCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewCometCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
