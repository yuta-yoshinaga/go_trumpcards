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

func mustCostlyColoursOutputJSON(msg string) string {
	out := &controller.CostlyColoursWebOutput{
		Players:       []*controller.CostlyColoursWebOutputPlayer{},
		Pile:          []*controller.WebOutputCard{},
		PlayableIdxs:  []int{},
		WentOut:       -1,
		WinnerIdx:     -1,
		HintHandIdx:   -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCostlyColoursOutputJSON: %v", err))
	}
	return string(b)
}

func TestCostlyColoursWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"mog"}`

	diMock := new(usecase.MockCostlyColoursInteractor)
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("Mog", mock.Anything).Return(mockOutput)
	diMock.On("Play", mock.Anything).Return(mockOutput)
	diMock.On("NextDeal").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewCostlyColoursWebController(func() uc.CostlyColoursInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.CostlyColoursWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustCostlyColoursOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultCostlyColoursConfig())
	})

	// **応じるか断るかは要求に乗せる。** 既定を「応じる」にすると、断るつもりの
	// 要求が黙って交換になる。
	t.Run("mog carries the decision", func(t *testing.T) {
		run(t, `{"command":"mog","accept":true,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Mog", true)
		run(t, `{"command":"mog","accept":false,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Mog", false)
	})
	t.Run("mog without a decision is refused", func(t *testing.T) {
		var input controller.CostlyColoursWebInput
		_ = json.Unmarshal([]byte(`{"command":"mog","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("play carries the hand index", func(t *testing.T) {
		run(t, `{"command":"play","handIndex":2,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 2)
	})
	// **0 は「省略」ではない。**
	t.Run("play without a hand index is refused", func(t *testing.T) {
		var input controller.CostlyColoursWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("nextdeal", func(t *testing.T) {
		run(t, `{"command":"nd","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "NextDeal")
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestCostlyColoursWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name       string
		in         controller.CostlyColoursWebConfig
		wantTarget int
	}{
		{"defaults", controller.CostlyColoursWebConfig{}, domain.CostlyColoursDefaultTarget},
		// **Parlett の 121 も選べる。**
		{"parlett target", controller.CostlyColoursWebConfig{TargetScore: i(121)}, 121},
		{"too high falls back", controller.CostlyColoursWebConfig{TargetScore: i(999)},
			domain.CostlyColoursDefaultTarget},
		{"too low falls back", controller.CostlyColoursWebConfig{TargetScore: i(1)},
			domain.CostlyColoursDefaultTarget},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantTarget, cfg.TargetScore)
			assert.NoError(t, cfg.Validate())
		})
	}
	// **既定は Cotton の 61 点。**
	assert.Equal(t, 61, domain.DefaultCostlyColoursConfig().TargetScore)
	wc := controller.CostlyColoursWebConfig{CpuDifficulty: i(9)}
	assert.NoError(t, wc.ToConfig().Validate(), "範囲外の難易度が丸められていない")
	assert.Equal(t, domain.DefaultCostlyColoursConfig(), controller.CostlyColoursWebInput{}.ToConfig())
}

func TestCostlyColoursCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"mog"}`

	newMock := func() *usecase.MockCostlyColoursInteractor {
		m := new(usecase.MockCostlyColoursInteractor)
		m.On("GetConfig").Return(domain.DefaultCostlyColoursConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Mog", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextDeal").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCostlyColoursCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCostlyColoursConfig())
	})

	// **応じる手と断る手は別コマンド。** 引数の有無で分けると、打ち間違いが
	// 黙って交換になる。
	for _, tt := range []struct {
		cmd    string
		accept bool
	}{
		{"mog", true}, {"m", true}, {"nomog", false}, {"nm", false},
	} {
		t.Run(tt.cmd+" sends the decision", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec(tt.cmd))
			m.AssertCalled(t, "Mog", tt.accept)
			m.AssertNotCalled(t, "Mog", !tt.accept)
		})
	}

	for _, cmd := range []string{"p 1", "play 1"} {
		t.Run(cmd+" plays a card", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec(cmd))
			m.AssertCalled(t, "Play", 1)
		})
	}

	t.Run("play without a hand index is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewCostlyColoursCuiController(m).Exec("p"),
			msgStem("cardIndexRequired"))
		assert.Contains(t, controller.NewCostlyColoursCuiController(m).Exec("p zz"),
			msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Play", mock.Anything)
	})

	for _, cmd := range []string{"nextdeal", "nd"} {
		t.Run(cmd+" advances the deal", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextDeal")
		})
	}

	t.Run("setdifficulty and settarget reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultCostlyColoursConfig()
		cfg.CpuDifficulty = domain.CostlyColoursCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m2).Exec("st 121"))
		cfg2 := domain.DefaultCostlyColoursConfig()
		cfg2.TargetScore = 121
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **勝負点は 31〜121。** 断る文言も範囲を名指す。
	t.Run("a target outside 31-121 is refused and names the range", func(t *testing.T) {
		m := newMock()
		out := controller.NewCostlyColoursCuiController(m).Exec("st 5")
		assert.Contains(t, out, msgStem("costlycolours.invalidTargetScore"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CostlyColoursMinTarget))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CostlyColoursMaxTarget))
		assert.Contains(t, controller.NewCostlyColoursCuiController(m).Exec("st 999"),
			msgStem("costlycolours.invalidTargetScore"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewCostlyColoursCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewCostlyColoursCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
